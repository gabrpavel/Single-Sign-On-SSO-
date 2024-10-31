package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/config"
	"sso/internal/services/auth"
	"sso/internal/storage/sso-db"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	cfg *config.Config,
) *App {

	storage, err := sso_db.New(cfg)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, cfg.TokenTTL)
	grpcApp := grpcapp.New(log, authService, cfg.GRPC.Port)

	return &App{
		GRPCSrv: grpcApp,
	}
}
