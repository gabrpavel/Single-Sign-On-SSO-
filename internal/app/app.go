package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/config"
	"sso/internal/services/auth"
	"sso/internal/storage"
	"sso/internal/storage/sso_db"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	cfg *config.Config,
) *App {

	db, err := sso_db.New(cfg)
	if err != nil {
		panic(err)
	}

	tokenStorage := storage.NewRedisTokenStore("localhost:6379", "", 0)

	authService := auth.New(log, db, db, db, tokenStorage, cfg.TokenTTL)
	grpcApp := grpcapp.New(log, authService, cfg.GRPC.Port)

	return &App{
		GRPCSrv: grpcApp,
	}
}
