package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"url-shortener/internal/lib/jwt"
	"url-shortener/internal/lib/logger/sl"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrConvert          = errors.New("failed to convert")
	ErrPermissionDenied = errors.New("don't have permission to action")
	ErrFailedAdminCheck = errors.New("failed to check if user is admin")
)

type Key string

var (
	authErrorKey = Key("authError")
	isAdminKey   = Key("isAdmin")
)

type PermissionProvider interface {
	IsAdmin(ctx context.Context, email string) (bool, error)
}

func New(
	log *slog.Logger,
	userKey string,
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

			token, err := jwt.Parse(jwtToken, userKey)
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
				slog.Bool("Is admin", token.Level > 1),
			)

			isAdmin, err := permProvider.IsAdmin(r.Context(), token.Email)
			if err != nil {
				log.Error("failed to check if user is admin", sl.Err(err))

				ctx := context.WithValue(r.Context(), authErrorKey, true)
				ctx = context.WithValue(ctx, isAdminKey, false)
				next.ServeHTTP(w, r.WithContext(ctx))

				return
			}

			entry.Info("user authorized")

			ctx := context.WithValue(r.Context(), isAdminKey, isAdmin)
			ctx = context.WithValue(ctx, authErrorKey, false)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

func CheckPermission(ctx context.Context) error {
	const op = "middleware.auth.CheckPermission"

	isAdmin, ok := ctx.Value(isAdminKey).(bool)
	if !ok {
		return fmt.Errorf("%s: %w", op, ErrConvert)
	}
	if !isAdmin {
		isErr, ok := ctx.Value(authErrorKey).(bool)
		if !ok {
			return fmt.Errorf("%s: %w", op, ErrConvert)
		}
		if isErr {
			return fmt.Errorf("%s: %w", op, ErrFailedAdminCheck)
		}
		return fmt.Errorf("%s: %w", op, ErrPermissionDenied)
	}

	return nil
}

func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	splitToken := strings.Split(authHeader, " ")
	if len(splitToken) != 2 {
		return ""
	}

	return splitToken[1]
}
