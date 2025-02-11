package postgres

import (
	"github.com/22Fariz22/forum/internal/model"
	"github.com/22Fariz22/forum/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Migrate applies database migrations
func Migrate(logger logger.Logger, dsn string) error {
	// Инициализация GORM с использованием только для миграций
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Debugf("Error in pkg/db/migrate/migrate.go")
		return err
	}

	// Выполнение миграций
	return db.AutoMigrate(&model.User{}, &model.Post{}, &model.Comment{})
}
