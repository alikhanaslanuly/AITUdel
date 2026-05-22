package grpc

import (
	"context"
	"user-service/internal/domain"
	"user-service/internal/usecase"
	pb "user-service/proto/user"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	authUC    *usecase.AuthUsecase
	notifUC   *usecase.NotifUsecase
	courierUC *usecase.CourierUsecase
}

func NewUserHandler(
	authUC *usecase.AuthUsecase,
	notifUC *usecase.NotifUsecase,
	courierUC *usecase.CourierUsecase,
) *UserHandler {
	return &UserHandler{
		authUC:    authUC,
		notifUC:   notifUC,
		courierUC: courierUC,
	}
}

func (h *UserHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	role := req.Role
	if role == "" {
		role = domain.RoleUser
	}

	user, accessToken, refreshToken, err := h.authUC.Register(
		ctx, req.Email, req.Password, req.Name, req.Phone, role,
	)
	if err != nil {
		return nil, err
	}

	go func() {
		notifType := domain.NotifWelcome
		_ = h.notifUC.Send(context.Background(), user.ID, notifType, "", "")
	}()

	return &pb.RegisterResponse{
		User:         userToProto(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *UserHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, accessToken, refreshToken, err := h.authUC.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	return &pb.LoginResponse{
		User:         userToProto(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *UserHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	accessToken, newRefresh, err := h.authUC.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &pb.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefresh,
	}, nil
}

func (h *UserHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if err := h.authUC.Logout(ctx, req.RefreshToken); err != nil {
		return &pb.LogoutResponse{Success: false}, nil
	}
	return &pb.LogoutResponse{Success: true}, nil
}

func (h *UserHandler) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.GetProfileResponse, error) {
	user, err := h.authUC.GetProfile(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &pb.GetProfileResponse{User: userToProto(user)}, nil
}

func (h *UserHandler) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	user, err := h.authUC.UpdateProfile(ctx, req.UserId, req.Name, req.Phone)
	if err != nil {
		return nil, err
	}
	return &pb.UpdateProfileResponse{User: userToProto(user)}, nil
}

func (h *UserHandler) SendNotification(ctx context.Context, req *pb.SendNotificationRequest) (*pb.SendNotificationResponse, error) {
	err := h.notifUC.Send(ctx, req.UserId, req.Type, req.Subject, req.Body)
	if err != nil {
		return &pb.SendNotificationResponse{Success: false}, nil
	}
	return &pb.SendNotificationResponse{Success: true}, nil
}

func (h *UserHandler) GetCourier(ctx context.Context, req *pb.GetCourierRequest) (*pb.GetCourierResponse, error) {
	c, err := h.courierUC.GetCourier(ctx, req.CourierId)
	if err != nil {
		return nil, err
	}
	return &pb.GetCourierResponse{Courier: courierToProto(c)}, nil
}

func (h *UserHandler) UpdateCourierStatus(ctx context.Context, req *pb.UpdateCourierStatusRequest) (*pb.UpdateCourierStatusResponse, error) {
	err := h.courierUC.UpdateStatus(ctx, req.CourierId, req.Status, req.Latitude, req.Longitude)
	if err != nil {
		return &pb.UpdateCourierStatusResponse{Success: false}, nil
	}
	return &pb.UpdateCourierStatusResponse{Success: true}, nil
}

func userToProto(u *domain.User) *pb.User {
	return &pb.User{
		Id:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		IsStudent: u.IsStudent,
		Phone:     u.Phone,
		CreatedAt: u.CreatedAt.String(),
	}
}

func courierToProto(c *domain.Courier) *pb.Courier {
	return &pb.Courier{
		Id:        c.ID,
		UserId:    c.UserID,
		Status:    c.Status,
		Latitude:  c.Latitude,
		Longitude: c.Longitude,
	}
}
