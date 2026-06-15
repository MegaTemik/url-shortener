package redirect

import (
	"errors"
	"log/slog"
	"net/http"

	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type URLRedirect interface {
	RegisterGetURL(alias string) (string, error)
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=URLRedirect
func New(log *slog.Logger, urlRedirect URLRedirect) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		resURL, err := urlRedirect.RegisterGetURL(alias)

		if err != nil {
			if errors.Is(err, service.ErrURLNotFound) {
				log.Info("url not found", "alias", alias)
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, resp.Error("url not found"))

				return
			}
		}

		if err != nil {
			log.Error("failed to get url", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("got url", slog.String("url", resURL))

		http.Redirect(w, r, resURL, http.StatusFound)
	}
}
