package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"url-shortener/internal/lib/jwt"
	"url-shortener/internal/lib/logger/sl"
)

var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrFailedIsAdminCheck = errors.New("failed to check if user is admin")
)

type Key string

var (
	authErrorKey = Key("authError")
	uIDKey       = Key("uid")
	isAdminKey   = Key("isAdmin")
)

type PermissionProvider interface {
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

func New(
	log *slog.Logger,
	appSecret string,
	permProvider PermissionProvider,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log := log.With(
			slog.String("component", "middleware/auth"),
		)

		log.Info("auth middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			jwtToken := extractBearerToken(r)
			if jwtToken == "" {
				next.ServeHTTP(w, r)
				return
			}

			token, err := jwt.Parse(jwtToken, appSecret)
			if err != nil {
				log.Warn("failed to parse token", sl.Err(err))

				ctx := context.WithValue(r.Context(), authErrorKey, ErrInvalidToken)
				next.ServeHTTP(w, r.WithContext(ctx))

				return
			}

			entry := log.With(
				slog.Int64("UID", token.UID),
				slog.String("Email", token.Email),
				slog.String("Expiration", token.Expiration.Format("15:05:05.000")),
				slog.Int64("AppID", token.AppID),
			)

			entry.Info("user authorized")

			isAdmin, err := permProvider.IsAdmin(r.Context(), token.UID)
			if err != nil {
				log.Error("failed to check if user is admin", sl.Err(err))

				ctx := context.WithValue(r.Context(), authErrorKey, ErrFailedIsAdminCheck)
				next.ServeHTTP(w, r.WithContext(ctx))

				return
			}

			ctx := context.WithValue(r.Context(), uIDKey, token.UID)
			ctx = context.WithValue(ctx, isAdminKey, isAdmin)

			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

func UIDFromContext(ctx context.Context) (int64, bool) {
	uid, ok := ctx.Value(uIDKey).(int64)
	return uid, ok
}

func ErrorFromContext(ctx context.Context) (error, bool) {
	err, ok := ctx.Value(authErrorKey).(error)
	return err, ok
}

func IsAdminFromContext(ctx context.Context) (bool, bool) {
	isAdmin, ok := ctx.Value(isAdminKey).(bool)
	return isAdmin, ok
}

func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	splitToken := strings.Split(authHeader, " ")
	if len(splitToken) != 2 {
		return ""
	}

	return splitToken[1]
}
