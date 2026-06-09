package update

import (
	"errors"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Alias  string `json:"alias" validate:"required"`
	NewURL string `json:"url" validate:"required,url"`
}

type Response struct {
	resp.Response
	CountUpdated int64 `json:"countUpdated"`
}

type UpdateURL interface {
	UpdateURL(alias, newURL string) (int64, error)
}

func New(log *slog.Logger, updateURL UpdateURL) http.HandlerFunc {
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
			render.JSON(w, r, resp.Error("failed to decode request bidy"))
			return
		}

		log.Info("request body decode", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("validation error", sl.Err(err))
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		alias := req.Alias
		newURL := req.NewURL

		countUpdated, err := updateURL.UpdateURL(alias, newURL)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)
			render.JSON(w, r, resp.Error("url not found"))
			return
		}

		if err != nil {
			log.Error("failed to update url", "alias", alias, sl.Err(err))
			render.JSON(w, r, resp.Error("failed to update url"))
			return
		}

		log.Info("url updated", slog.String("alias", alias), slog.Int64("count_updated", countUpdated))

		responseOK(w, r, countUpdated)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, countUpdated int64) {
	render.JSON(w, r, Response{
		Response:     resp.OK(),
		CountUpdated: countUpdated,
	})
}
