package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/hr0mk4/grpc_auth/internal/domain/models"
	"github.com/hr0mk4/grpc_auth/internal/lib/jwt"
	"github.com/hr0mk4/grpc_auth/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
)

type Auth struct {
	log         *slog.Logger
	userSaver   UserSaver
	userGetter  UserGetter
	userChecker UserChecker
	appGetter   AppGetter
	tokenTTL    time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserGetter interface {
	GetUser(ctx context.Context, email string) (user models.User, err error)
}

type UserChecker interface {
	IsAdmin(ctx context.Context, uid int64) (isAdmin bool, err error)
}

type AppGetter interface {
	GetApp(ctx context.Context, id int32) (app models.App, err error)
}

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userGetter UserGetter,
	userChecker UserChecker,
	appGetter AppGetter,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:         log,
		userSaver:   userSaver,
		userGetter:  userGetter,
		userChecker: userChecker,
		appGetter:   appGetter,
		tokenTTL:    tokenTTL,
	}
}

// Login checks if the user with given email and password exists
// If it does, it returns a token.
// If not, it returns an error
func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appId int32,
) (string, error) {
	const op = "auth.Login"

	log := a.log.With(slog.String("op", op), slog.String("email", email))

	log.Info("logging in user")

	user, err := a.userGetter.GetUser(ctx, email)

	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", slog.String("error", err.Error()))
			return "", fmt.Errorf("%s %w", op, ErrInvalidCredentials)
		}

		log.Error("failed to get user", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.HashPass, []byte(password)); err != nil {
		log.Info("invalid password", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s %w", op, ErrInvalidCredentials)
	}

	app, err := a.appGetter.GetApp(ctx, appId)

	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found", slog.String("error", err.Error()))
			return "", fmt.Errorf("%s %w", op, ErrInvalidCredentials)
		}
		log.Error("failed to get app", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s %w", op, err)
	}

	token, err := jwt.GetToken(user, app, a.tokenTTL)

	if err != nil {
		log.Error("failed to get token", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s %w", op, err)
	}

	log.Info("user logged in", slog.String("app", app.Name))

	return token, nil
}

// Register checks if the user with given email and password already exists.
// If it does, it returns an error.
// If not, it creates a new user and returns an UID.
func (a *Auth) Register(
	ctx context.Context,
	email string,
	password string,
) (int64, error) {
	const op = "auth.Register"

	log := a.log.With(slog.String("op", op), slog.String("email", email))

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		log.Error("failed to hash password", slog.String("error", err.Error()))
		return 0, fmt.Errorf("%s %w", op, err)
	}

	id, err := a.userSaver.SaveUser(ctx, email, passHash)

	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", slog.String("error", err.Error()))
			return 0, fmt.Errorf("%s %w", op, ErrUserExists)
		}
		log.Error("failed to save user", slog.String("error", err.Error()))
		return 0, fmt.Errorf("%s %w", op, err)
	}

	log.Info("user registered")

	return id, nil
}

// IsAdmin checks if the user with given UID is admin
func (a *Auth) IsAdmin(
	ctx context.Context,
	userId int64,
) (bool, error) {
	const op = "auth.IsAdmin"

	log := a.log.With(slog.String("op", op), slog.Int64("uid", userId))

	log.Info("checking if user is admin")

	isAdmin, err := a.userChecker.IsAdmin(ctx, userId)

	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", slog.String("error", err.Error()))
			return false, fmt.Errorf("%s %w", op, ErrUserNotFound)
		}
		log.Error("failed to check if user is admin", slog.String("error", err.Error()))
		return false, fmt.Errorf("%s %w", op, err)
	}

	log.Info("user was checked if admin", slog.Bool("isAdmin", isAdmin))

	return isAdmin, nil
}
