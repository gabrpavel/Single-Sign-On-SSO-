package auth

import (
	"context"
	"errors"
	ssov2 "github.com/gabrpavel/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sso/internal/services/auth"
)

type Auth interface {
	Login(ctx context.Context,
		email string,
		password string,
		appID int,
	) (token string, err error)
	RegisterNewUser(ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	VerifyToken(ctx context.Context,
		token string,
	) (bool, int64, string, error)
}

type serverAPI struct {
	ssov2.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov2.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

const (
	emptyValue = 0
)

func (s *serverAPI) Login(
	ctx context.Context,
	req *ssov2.LoginRequest,
) (*ssov2.LoginResponse, error) {

	if err := validateLogin(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}

		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &ssov2.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *ssov2.RegisterRequest,
) (*ssov2.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	userID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &ssov2.RegisterResponse{UserId: userID}, nil
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *ssov2.IsAdminRequest,
) (*ssov2.IsAdminResponse, error) {
	if err := validateIsAdmin(req); err != nil {
		return nil, err
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov2.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

func (s *serverAPI) VerifyToken(
	ctx context.Context,
	req *ssov2.VerifyTokenRequest,
) (*ssov2.VerifyTokenResponse, error) {
	// Проверка, что токен не пустой
	if req.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	// Вызов метода VerifyToken у сервиса auth
	_, userID, email, err := s.auth.VerifyToken(ctx, req.GetToken())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidToken) {
			return &ssov2.VerifyTokenResponse{IsValid: false}, nil
		}
		return nil, status.Error(codes.Internal, "failed to verify token")
	}

	// Если токен валидный, возвращаем данные пользователя
	return &ssov2.VerifyTokenResponse{
		IsValid: true,
		UserId:  userID,
		Email:   email,
	}, nil
}

func validateLogin(req *ssov2.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	if req.GetAppId() == emptyValue {
		return status.Error(codes.InvalidArgument, "app_id is required")
	}

	return nil
}

func validateRegister(req *ssov2.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	return nil
}

func validateIsAdmin(req *ssov2.IsAdminRequest) error {
	if req.GetUserId() == emptyValue {
		return status.Error(codes.Internal, "user_id is required")
	}

	return nil
}
