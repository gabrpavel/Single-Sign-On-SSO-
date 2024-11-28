package auth

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/jwt"
	"sso/internal/lib/logger/sl"
	"sso/internal/storage"
	"time"
)

type Auth struct {
	log          *slog.Logger
	usrSaver     UserSaver
	usrProvider  UserProvider
	appProvider  AppProvider
	tokenStorage storage.TokenStorage
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, uid int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid token")
)

// New returns a new instance of the Auth service.
func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenStorage storage.TokenStorage,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		usrSaver:     userSaver,
		usrProvider:  userProvider,
		log:          log,
		appProvider:  appProvider,
		tokenStorage: tokenStorage,
		tokenTTL:     tokenTTL}
}

// Login checks if user with given credentials exists in the system and returns access token.
//
// If user exists, but password is incorrect, returns error.
// If user doesn't exist, returns error.
func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (string, error) {
	const op = "Auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("username", email),
	)
	log.Info("attempting to login user")

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		a.log.Error("failed to get user", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in successfully")

	token, err := jwt.NewAuthToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := a.tokenStorage.SaveToken(ctx, user.ID, token, a.tokenTTL); err != nil {
		a.log.Error("failed to save token in Redis", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

// RegisterNewUser registers new user in the system and returns user ID.
// If user with given username already exists, returns error.
func (a *Auth) RegisterNewUser(ctx context.Context, email string, pass string) (int64, error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", sl.Err(err))

			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}

		log.Error("failed to save user", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered")

	return id, nil
}

// IsAdmin checks if user is admin.
func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "Auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
	)

	log.Info("checking if user is admin")

	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("user not found", sl.Err(err))

			return false, fmt.Errorf("%s: %w", op, ErrInvalidAppID)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}

func (a *Auth) VerifyToken(ctx context.Context, token string) (bool, int64, string, error) {
	const op = "Auth.VerifyToken"

	log := a.log.With(
		slog.String("op", op),
		slog.String("token", token),
	)

	log.Info("verifying token")

	claims, err := jwt.ParseToken(token, func(appID int) (string, error) {
		app, err := a.appProvider.App(ctx, appID)
		if err != nil {
			return "", err
		}
		return app.Secret, nil
	})
	if err != nil {
		log.Warn("invalid token", sl.Err(err))
		return false, 0, "", fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	userID, ok := claims["uid"].(float64)
	if !ok {
		log.Warn("error while parsing token", sl.Err(err))
		return false, 0, "", fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	email, ok := claims["email"].(string)
	if !ok {
		log.Warn("error while parsing token", sl.Err(err))
		return false, 0, "", fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	log.Info("verifying token", slog.Bool("isValid", true))

	return true, int64(userID), email, nil
}
