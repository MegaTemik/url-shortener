package save

import (
	"errors"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/service"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Request struct {
	URL   string `json:"url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=URLSaver
type URLSaver interface {
	RegisterSaveURL(inputReq service.CreateURLInput) (string, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		inputReq := service.CreateURLInput{
			URL:   req.URL,
			Alias: req.Alias,
		}

		alias, err := urlSaver.RegisterSaveURL(inputReq)
		if err != nil {
			if errors.Is(err, service.ErrInvalidURL) {
				log.Error("invalid url", sl.Err(err))
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.Error("invalid url"))
				return

			} else if errors.Is(err, service.ErrURLExists) {
				log.Info("url already exists", slog.String("url", req.URL))
				render.Status(r, http.StatusConflict)
				render.JSON(w, r, resp.Error("url already exists"))
				return
			}
		}

		if err != nil {
			log.Error("failed to add url", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to add url"))

			return
		}

		log.Info("url added", slog.String("alias", alias))

		render.Status(r, http.StatusCreated)
		responseOK(w, r, alias)

	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}
