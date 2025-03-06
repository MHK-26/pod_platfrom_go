package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/your-username/podcast-platform/pkg/auth/models"
	"github.com/your-username/podcast-platform/pkg/auth/usecase"
	pb "github.com/your-username/podcast-platform/api/proto/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler is the gRPC handler for the auth service
type Handler struct {
	pb.UnimplementedAuthServiceServer
	usecase usecase.Usecase
}

// NewHandler creates a new auth gRPC handler
func NewHandler(usecase usecase.Usecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

// VerifyToken verifies a token and returns the user ID
func (h *Handler) VerifyToken(ctx context.Context, req *pb.VerifyTokenRequest) (*pb.VerifyTokenResponse, error) {
	payload, err := h.usecase.VerifyToken(ctx, req.Token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid token: %v", err)
	}

	return &pb.VerifyTokenResponse{
		UserId:   payload.UserID.String(),
		Email:    payload.Email,
		UserType: payload.UserType,
	}, nil
}

// GetUserByID gets a user by ID
func (h *Handler) GetUserByID(ctx context.Context, req *pb.GetUserByIDRequest) (*pb.User, error) {
	userID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid user ID: %v", err)
	}

	user, err := h.usecase.GetUserByID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "User not found: %v", err)
	}

	return convertUserToProto(user), nil
}

// Helper function to convert user model to proto
func convertUserToProto(user *models.User) *pb.User {
	pbUser := &pb.User{
		Id:                user.ID.String(),
		Email:             user.Email,
		Username:          user.Username,
		FullName:          user.FullName,
		Bio:               user.Bio,
		ProfileImageUrl:   user.ProfileImageURL,
		UserType:          user.UserType,
		AuthProvider:      user.AuthProvider,
		AuthProviderId:    user.AuthProviderID,
		IsVerified:        user.IsVerified,
		PreferredLanguage: user.PreferredLanguage,
		CreatedAt:         timestamppb.New(user.CreatedAt),
		UpdatedAt:         timestamppb.New(user.UpdatedAt),
	}

	if user.LastLoginAt != nil {
		pbUser.LastLoginAt = timestamppb.New(*user.LastLoginAt)
	}

	return pbUser
}