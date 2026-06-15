package update

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
	Alias  string `json:"alias"`
	NewURL string `json:"url"`
}

type Response struct {
	resp.Response
	CountUpdated int64 `json:"countUpdated"`
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=URLUpdater
type URLUpdater interface {
	RegisterUpdateURL(inputReq service.UpdateURLInput) (int64, error)
}

func New(log *slog.Logger, urlUpdate URLUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.update.New"

		log = log.With(
			slog.String("fn", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failded to decode request body", sl.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		log.Info("request body decode", slog.Any("request", req))

		inputReq := service.UpdateURLInput{
			Alias:  req.Alias,
			NewURL: req.NewURL,
		}

		countUpdated, err := urlUpdate.RegisterUpdateURL(inputReq)
		if err != nil {
			if errors.Is(err, service.ErrInvalidURL) {
				log.Error("invalid url", sl.Err(err))
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.Error("invalid url"))
				return
			} else if errors.Is(err, service.ErrURLNotFound) {
				log.Info("url not found", "alias", inputReq.Alias)
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, resp.Error("url not found"))
				return
			}
		}
		if err != nil {
			log.Error("failed to update url", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to update url"))
			return
		}

		log.Info("url updated", slog.String("alias", inputReq.Alias), slog.Int64("count_updated", countUpdated))

		responseOK(w, r, countUpdated)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, countUpdated int64) {
	render.JSON(w, r, Response{
		Response:     resp.OK(),
		CountUpdated: countUpdated,
	})
}
