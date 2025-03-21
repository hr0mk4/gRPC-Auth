package auth

import (
	"context"
	"errors"
	"net/mail"

	"github.com/hr0mk4/grpc_auth/internal/services/auth"
	authv1 "github.com/hr0mk4/protos_auth/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	emptyValue = 0
)

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

type serverAPI struct {
	authv1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	authv1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *authv1.RegisterRequest,
) (*authv1.RegisterResponse, error) {
	err := validateRegister(req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	userid, err := s.auth.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(
				codes.AlreadyExists,
				"user already exists",
			)
		}
		return nil, status.Error(
			codes.Internal,
			"internal error",
		)
	}

	return &authv1.RegisterResponse{
		UserId: userid,
	}, nil
}

func (s *serverAPI) LogIn(
	ctx context.Context,
	req *authv1.LoginRequest,
) (*authv1.LoginResponse, error) {
	err := validateLogin(req.GetEmail(), req.GetPassword(), req.GetAppId())
	if err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetAppId())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(
				codes.InvalidArgument,
				"invalid credentials",
			)
		}
		return nil, status.Error(
			codes.Internal,
			"internal error",
		)
	}

	return &authv1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *authv1.IsAdminRequest,
) (*authv1.IsAdminResponse, error) {
	err := validateIsAdmin(req.GetUserId())
	if err != nil {
		return nil, err
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(
				codes.NotFound,
				"user not found",
			)
		}
		return nil, status.Error(
			codes.Internal,
			"internal error",
		)
	}

	return &authv1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

func validateLogin(
	email string,
	password string,
	appId int32,
) error {
	if email == "" || password == "" {
		return status.Error(
			codes.InvalidArgument,
			"both email and password are required",
		)
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return status.Error(
			codes.InvalidArgument,
			"invalid email",
		)
	}

	if appId == emptyValue {
		return status.Errorf(
			codes.InvalidArgument,
			"app_id is required",
		)
	}
	return nil
}

func validateRegister(
	email string,
	password string,
) error {
	if email == "" || password == "" {
		return status.Error(
			codes.InvalidArgument,
			"both email and password are required",
		)
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return status.Error(
			codes.InvalidArgument,
			"invalid email",
		)
	}
	return nil
}

func validateIsAdmin(
	userId int64,
) error {
	if userId == emptyValue {
		return status.Errorf(
			codes.InvalidArgument,
			"user_id is required",
		)
	}
	return nil
}
