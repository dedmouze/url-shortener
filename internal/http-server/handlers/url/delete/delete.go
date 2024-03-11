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

		err := auth.CheckPermission(r.Context())
		if err != nil {
			if errors.Is(err, auth.ErrConvert) {
				log.Error("failed to convert", sl.Err(err))
				render.JSON(w, r, response.Error("failed to convert token"))
			} else if errors.Is(err, auth.ErrFailedAdminCheck) {
				log.Error("failed to check if user is admin", sl.Err(err))
				render.JSON(w, r, response.Error("failed to check if user is admin"))
			} else if errors.Is(err, auth.ErrInvalidToken) {
				log.Error("invalid token", sl.Err(err))
				render.JSON(w, r, response.Error("invalid token"))
			} else if errors.Is(err, auth.ErrPermissionDenied) {
				log.Error("don't have permission to action", sl.Err(err))
				render.JSON(w, r, response.Error("don't have permission to action"))
			} else {
				log.Error("internal error", sl.Err(err))
				render.JSON(w, r, response.Error("internal error"))
			}
			return
		}

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Error("alias is empty")
			render.JSON(w, r, response.Error("invalid request"))
			return
		}

		err = urldeleter.DeleteURL(alias)
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
