package grpcApp

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	authgRPC "github.com/hr0mk4/grpc_auth/internal/grpc/auth"
	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appId int32,
	) (token string, err error)
	Register(
		ctx context.Context,
		email string,
		password string,
	) (userId int64, err error)
	IsAdmin(
		ctx context.Context,
		userId int64,
	) (isAdmin bool, err error)
}

func New(log *slog.Logger, port int, authService Auth) *App {
	gRPCServer := grpc.NewServer()

	authgRPC.Register(gRPCServer, authService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcApp.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		log.Error("failed to listen", err)
		return err
	}
	log.Info("starting gRPC server", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		log.Error("failed to serve", err)
		return err
	}

	return nil
}

func (a *App) Stop() error {
	const op = "grpcApp.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()

	return nil
}
