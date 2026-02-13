package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/Egor-Pomidor-pdf/EventBooker/internal/config"
	handlers "github.com/Egor-Pomidor-pdf/EventBooker/internal/handler"
	"github.com/Egor-Pomidor-pdf/EventBooker/internal/repository"
	"github.com/Egor-Pomidor-pdf/EventBooker/internal/service"
	postgres "github.com/Egor-Pomidor-pdf/EventBooker/pkg/db"
	"github.com/Egor-Pomidor-pdf/EventBooker/pkg/server"
	"github.com/robfig/cron/v3"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

func main() {

	// make context
	ctx := context.Background()
	ctx, ctxStop := signal.NotifyContext(ctx, os.Interrupt)

	// init config
	cfg, err := config.NewConfig("", "")
	if err != nil {
		log.Fatal(err)
	}

	// init logger
	zlog.InitConsole()
	err = zlog.SetLevel(cfg.Env)
	if err != nil {
		log.Fatal(fmt.Errorf("error setting log level to '%s': %w", cfg.Env, err))
	}
	zlog.Logger.Info().
		Str("env", cfg.Env).
		Msg("Start app...")

	// strategies
	postgresRetryStrategy := config.MakeStrategy(cfg.Database.PostgresRetryConfig)

	// connect to db
	var postgresDB *dbpg.DB
	err = retry.DoContext(ctx, postgresRetryStrategy, func() error {
		var postgresConnErr error
		postgresDB, postgresConnErr = dbpg.New(cfg.Database.MasterDSN, cfg.Database.SlaveDSNs,
			&dbpg.Options{
				MaxOpenConns:    cfg.Database.MaxOpenConnections,
				MaxIdleConns:    cfg.Database.MaxIdleConnections,
				ConnMaxLifetime: time.Duration(cfg.Database.ConnectionMaxLifetimeSeconds) * time.Second,
			})
		return postgresConnErr
	})
	defer postgresDB.Master.Close()

	if err != nil {
		zlog.Logger.Fatal().
			Err(err).
			Msg("failed to connect to database")
	}

	zlog.Logger.Info().Msg("Successfully connected to PostgreSQL")

	zlog.Logger.Info().
		Str("MasterDSN", cfg.Database.MasterDSN).
		Msg("Start app...")
	// create migrations db
	migrationsPathPostgress := "file:///app/internal/migrations"
	err = postgres.MigrateUp(cfg.Database.MasterDSN, migrationsPathPostgress)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("couldn't migrate postgres on master DSN")
	}

	zlog.Logger.Info().Msg("Successfully connected to create migrations fo PSQL")

	// init repo
	store := repository.NewPostgresRepository(postgresDB)

	// inint service
	srv := service.NewService(store)
	handl := handlers.NewHandker(srv)
	router := handlers.NewRouter(handl)

	// Запускаем фоновый процесс для очистки просроченных бронирований
	c := cron.New()
	_, err = c.AddFunc("@every 1m", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		cancelled, err := srv.CancelExpiredBookings(ctx)
		if err != nil {
			zlog.Logger.Error().
				Err(err).
				Msg("ошибка при отмене просроченных бронирований")
		} else if len(cancelled) > 0 {
			zlog.Logger.Info().
				Int("cancelled_count", len(cancelled)).
				Msg("отменены просроченные бронирования")
		}
	})
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("ошибка настройки cron")
	}
	c.Start()
	defer c.Stop()

	zlog.Logger.Info().Msg("Фоновый процесс очистки бронирований запущен (каждую минуту)")

	// running server
	zlog.Logger.Info().Msg("server start")
	httpServer := server.NewHTTPServer(router)
	err = httpServer.GracefulRun(ctx, cfg.Server.Host, cfg.Server.Port)

	if err != nil {
		zlog.Logger.Error().
			Err(err).
			Msg("failed GracefulRun server")
	}

	zlog.Logger.Info().Msg("server gracefully stopped")

	ctxStop()
	zlog.Logger.Info().Msg("background operations gracefully stopped")
}
