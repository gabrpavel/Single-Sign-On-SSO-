package sso_db

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sso/internal/config"
	"sso/internal/domain/models"
	"sso/internal/storage"
)

type Storage struct {
	db *gorm.DB
}

// New создает новый экземпляр PostgreSQL.
func New(config *config.Config) (*Storage, error) {
	const op = "storage.sso-db.New"

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		config.Db.Host, config.Db.User, config.Db.Password, config.Db.DBName, config.Db.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = db.AutoMigrate(&models.User{}, &models.App{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

// SaveUser сохраняет пользователя в базе данных.
func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "storage.sso-db.SaveUser"

	user := models.User{Email: email, PassHash: passHash}
	result := s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&user)
	if result.Error != nil {
		return 0, fmt.Errorf("%s: %w", op, result.Error)
	}

	if result.RowsAffected == 0 {
		return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
	}

	return user.ID, nil
}

// User возвращает пользователя по email.
func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.sso-db.User"

	var user models.User
	result := s.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
	} else if result.Error != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, result.Error)
	}

	return user, nil
}

// IsAdmin проверяет, является ли пользователь администратором.
func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.sso-db.IsAdmin"

	var user models.User
	result := s.db.WithContext(ctx).Select("is_admin").Where("id = ?", userID).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
	} else if result.Error != nil {
		return false, fmt.Errorf("%s: %w", op, result.Error)
	}

	return user.IsAdmin, nil
}

// App возвращает приложение по его ID.
func (s *Storage) App(ctx context.Context, id int) (models.App, error) {
	const op = "storage.sso-db.App"

	var app models.App
	result := s.db.WithContext(ctx).Where("id = ?", id).First(&app)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
	} else if result.Error != nil {
		return models.App{}, fmt.Errorf("%s: %w", op, result.Error)
	}

	return app, nil
}
