package main

import (
	"fmt"
	"log"

	"github.com/22Fariz22/forum/config"
	"github.com/22Fariz22/forum/graph"
	"github.com/22Fariz22/forum/internal/repository"
	"github.com/22Fariz22/forum/internal/server"
	"github.com/22Fariz22/forum/pkg/db/postgres"
	"github.com/22Fariz22/forum/pkg/logger"
	"github.com/jmoiron/sqlx"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	appLogger := logger.NewApiLogger(cfg)

	appLogger.InitLogger()
	appLogger.Infof("AppVersion:%s, LogLevel:%s, Mode:%s", cfg.Server.AppVersion, cfg.Logger.Level, cfg.Server.Mode)

	var repo graph.Repository
	var psqlDB *sqlx.DB

	if cfg.Storage.StorageType == "postgres" {
		appLogger.Infof("storage:postgres")
		appLogger.Debugf("Postgres config: host=%s port=%s user=%s dbname=%s sslmode=%s password=%s PgDriver=%s",
			cfg.Postgres.PostgresqlHost,
			cfg.Postgres.PostgresqlPort,
			cfg.Postgres.PostgresqlUser,
			cfg.Postgres.PostgresqlDbname,
			cfg.Postgres.PostgresqlSSLMode,
			cfg.Postgres.PostgresqlPassword,
			cfg.Postgres.PgDriver,
		)

		// Формирование строки подключения для GORM
		dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s",
			cfg.Postgres.PostgresqlHost,
			cfg.Postgres.PostgresqlPort,
			cfg.Postgres.PostgresqlUser,
			cfg.Postgres.PostgresqlDbname,
			cfg.Postgres.PostgresqlPassword,
		)

		// Выполнение миграций
		if err := postgres.Migrate(appLogger, dsn); err != nil {
			appLogger.Errorf("Failed to run migrations: %v", err)
		}
		appLogger.Debug("Database migrated successfully")

		psqlDB, err = postgres.NewPsqlDB(cfg)
		if err != nil {
			appLogger.Fatalf("Postgresql init: %s", err)
		} else {
			appLogger.Infof("Postgres connected, Status: %#v", psqlDB.Stats())
		}
		defer psqlDB.Close()

		repo, err = repository.NewPostgresRepository(psqlDB)
		if err != nil {
			fmt.Println("err in ping", err)
			return
		}

	} else {
		appLogger.Infof("storage:inmemory")

		repo = repository.NewInMemoryRepository()
	}

	// Инициализируем резолвер с хранилищем и системой pubsub для подписок
	resolver := graph.NewResolver(repo)

	s := server.NewServer(appLogger, cfg, resolver)
	s.Run() //сделать возврат ошибки
}
