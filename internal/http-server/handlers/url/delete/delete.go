package delete

import (
	"errors"
	"log/slog"
	"net/http"

	"url-shortener/internal/http-server/middleware/auth"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type URLDeleter interface {
	DeleteURL(alias string) error
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLDeleter
func New(log *slog.Logger, urldeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.new"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		isAdmin, ok := auth.IsAdminFromContext(r.Context())
		if !ok {
			log.Error("failed to convert")
			render.JSON(w, r, response.Error("failed to convert"))
			return
		}

		if !isAdmin {
			log.Warn("user not admin")
			render.JSON(w, r, response.Error("don't have permission to delete"))
			return
		}

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Error("alias is empty")
			render.JSON(w, r, response.Error("invalid request"))
			return
		}

		err := urldeleter.DeleteURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				log.Info("url not found")
			} else {
				log.Error("failed to delete url", sl.Err(err))
				render.JSON(w, r, response.Error("internal error"))
				return
			}
		}

		log.Info("url deleted")
		render.JSON(w, r, response.OK())
	}
}
