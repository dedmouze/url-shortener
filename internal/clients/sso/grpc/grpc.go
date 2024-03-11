package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"url-shortener/domain/models"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"

	ssov1 "github.com/dedmouze/protos/gen/go/sso"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	api     ssov1.UserInfoClient
	apiKey  string
	UserKey string
}

type ClientSaver interface {
	SaveClient(name, apiKey, userKey string) error
}

type ClientGetter interface {
	Client(name string) (models.Client, error)
}

// TODO: configure transport security
func New(
	ctx context.Context,
	log *slog.Logger,
	addr string,
	timeout time.Duration,
	retriesCount int,
	appName string,
	clientSaver ClientSaver,
	clientGetter ClientGetter,
) (*Client, error) {
	const op = "clients.sso.grpc.New"

	name := "SSO"

	isRegistered := true
	client, err := clientGetter.Client(name)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			isRegistered = false
		} else {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	apiKey := client.ApiKey
	userKey := client.UserKey
	if !isRegistered {
		cc, err := grpc.DialContext(ctx, addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()), //disabling transport security
			grpc.WithChainUnaryInterceptor(
				grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
				grpcretry.UnaryClientInterceptor(retryOpts...),
			),
		)

		if err != nil {
			log.Error("connection to sso service failed", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		apiKey, userKey, err = registerApp(ctx, cc, appName)
		if err != nil {
			log.Error("failed to register app", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	auth := &auth{apiKey: apiKey}
	cc, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), //disabling transport security
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		),
		grpc.WithPerRPCCredentials(auth),
	)

	if err != nil {
		log.Error("connection to sso service failed", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	//TODO: change name insertion
	err = clientSaver.SaveClient(name, apiKey, userKey)
	if err != nil {
		if !errors.Is(err, storage.ErrAppExists) {
			log.Error("failed to save client", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	grpcClient := ssov1.NewUserInfoClient(cc)

	return &Client{
		api:     grpcClient,
		apiKey:  apiKey,
		UserKey: userKey,
	}, nil
}

func (c *Client) IsAdmin(ctx context.Context, email string) (bool, error) {
	const op = "clients.sso.grpc.Admin"

	resp, err := c.api.Admin(ctx, &ssov1.AdminRequest{Email: email})
	if err != nil {
		if statusError, ok := status.FromError(err); ok {
			if statusError.Code() == codes.NotFound {
				return false, nil
			}
			return false, fmt.Errorf("%s: %w", op, err)
		}
	}

	return resp.Level > 1, nil
}

// InterceptorLogger adapts slog logger to interceptor logger
func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func registerApp(ctx context.Context, cc *grpc.ClientConn, name string) (string, string, error) {
	const op = "clients.sso.grpc.registerApp"

	resp, err := ssov1.NewAuthClient(cc).RegisterApp(ctx, &ssov1.RegisterAppRequest{Name: name})
	if err != nil {
		if statusError, ok := status.FromError(err); ok {
			if statusError.Code() != codes.AlreadyExists {
				return "", "", fmt.Errorf("%s: %w", op, err)
			}
		}
	}

	return resp.GetApiKey(), resp.GetUserKey(), nil
}

type auth struct {
	apiKey string
}

func (a auth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + a.apiKey,
	}, nil
}

func (a auth) RequireTransportSecurity() bool {
	return false
}
