package app

import (
	"log/slog"
	"time"

	grpcApp "github.com/hr0mk4/grpc_auth/internal/app/grpc"
	"github.com/hr0mk4/grpc_auth/internal/services/auth"
	"github.com/hr0mk4/grpc_auth/internal/storage/sqlite"
)

type App struct {
	GRPCServer *grpcApp.App
}

func New(
	log *slog.Logger,
	port int,
	storagePath string,
	tokenTTL time.Duration,
) *App {
	storage, err := sqlite.New(storagePath)

	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, storage, tokenTTL)

	grpcApp := grpcApp.New(log, port, authService)

	return &App{
		GRPCServer: grpcApp,
	}
}
