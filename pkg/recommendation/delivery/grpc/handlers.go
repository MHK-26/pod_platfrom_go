// pkg/recommendation/delivery/grpc/handlers.go
package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/MHK-26/pod_platfrom_go/pkg/recommendation/models"
	"github.com/MHK-26/pod_platfrom_go/pkg/recommendation/usecase"
	pb "github.com/MHK-26/pod_platfrom_go/api/proto/recommendation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler is the gRPC handler for the recommendation service
type Handler struct {
	pb.UnimplementedRecommendationServiceServer
	usecase usecase.Usecase
}

// NewHandler creates a new recommendation gRPC handler
func NewHandler(usecase usecase.Usecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

// GetPersonalizedRecommendations gets personalized recommendations for a user
func (h *Handler) GetPersonalizedRecommendations(ctx context.Context, req *pb.GetPersonalizedRecommendationsRequest) (*pb.GetRecommendationsResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid user ID: %v", err)
	}

	var excludedIDs []uuid.UUID
	for _, idStr := range req.ExcludedIds {
		id, err := uuid.Parse(idStr)
		if err == nil {
			excludedIDs = append(excludedIDs, id)
		}
	}

	// Prepare request
	modelReq := &models.RecommendationRequest{
		UserID:      userID,
		Limit:       int(req.Limit),
		ExcludedIDs: excludedIDs,
	}

	// Get recommendations
	response, err := h.usecase.GetPersonalizedRecommendations(ctx, modelReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get recommendations: %v", err)
	}

	// Convert to gRPC response
	return convertToGRPCResponse(response), nil
}

// GetSimilarPodcasts gets podcasts similar to a specified podcast
func (h *Handler) GetSimilarPodcasts(ctx context.Context, req *pb.GetSimilarPodcastsRequest) (*pb.GetRecommendationsResponse, error) {
	podcastID, err := uuid.Parse(req.PodcastId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid podcast ID: %v", err)
	}

	var excludedIDs []uuid.UUID
	for _, idStr := range req.ExcludedIds {
		id, err := uuid.Parse(idStr)
		if err == nil {
			excludedIDs = append(excludedIDs, id)
		}
	}

	// Prepare request
	modelReq := &models.SimilarContentRequest{
		ContentID:   podcastID,
		ContentType: "podcast",
		Limit:       int(req.Limit),
		ExcludedIDs: excludedIDs,
	}

	// Get similar podcasts
	response, err := h.usecase.GetSimilarPodcasts(ctx, modelReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get similar podcasts: %v", err)
	}

	// Convert to gRPC response
	return convertToGRPCResponse(response), nil
}

// GetSimilarEpisodes gets episodes similar to a specified episode
func (h *Handler) GetSimilarEpisodes(ctx context.Context, req *pb.GetSimilarEpisodesRequest) (*pb.GetRecommendationsResponse, error) {
	episodeID, err := uuid.Parse(req.EpisodeId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid episode ID: %v", err)
	}

	var excludedIDs []uuid.UUID
	for _, idStr := range req.ExcludedIds {
		id, err := uuid.Parse(idStr)
		if err == nil {
			excludedIDs = append(excludedIDs, id)
		}
	}

	// Prepare request
	modelReq := &models.SimilarContentRequest{
		ContentID:   episodeID,
		ContentType: "episode",
		Limit:       int(req.Limit),
		ExcludedIDs: excludedIDs,
	}

	// Get similar episodes
	response, err := h.usecase.GetSimilarEpisodes(ctx, modelReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get similar episodes: %v", err)
	}

	// Convert to gRPC response
	return convertToGRPCResponse(response), nil
}

// GetTrendingPodcasts gets trending podcasts
func (h *Handler) GetTrendingPodcasts(ctx context.Context, req *pb.GetTrendingPodcastsRequest) (*pb.GetRecommendationsResponse, error) {
	var excludedIDs []uuid.UUID
	for _, idStr := range req.ExcludedIds {
		id, err := uuid.Parse(idStr)
		if err == nil {
			excludedIDs = append(excludedIDs, id)
		}
	}

	// Prepare request
	modelReq := &models.TrendingRequest{
		TimeRange:   req.TimeRange,
		Limit:       int(req.Limit),
		ExcludedIDs: excludedIDs,
	}

	// Get trending podcasts
	response, err := h.usecase.GetTrendingPodcasts(ctx, modelReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get trending podcasts: %v", err)
	}

	// Convert to gRPC response
	return convertToGRPCResponse(response), nil
}

// GetPopularInCategory gets popular podcasts in a category
func (h *Handler) GetPopularInCategory(ctx context.Context, req *pb.GetPopularInCategoryRequest) (*pb.GetRecommendationsResponse, error) {
	categoryID, err := uuid.Parse(req.CategoryId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid category ID: %v", err)
	}

	var excludedIDs []uuid.UUID
	for _, idStr := range req.ExcludedIds {
		id, err := uuid.Parse(idStr)
		if err == nil {
			excludedIDs = append(excludedIDs, id)
		}
	}

	// Prepare request
	modelReq := &models.CategoryPopularRequest{
		CategoryID:  categoryID,
		Limit:       int(req.Limit),
		ExcludedIDs: excludedIDs,
	}

	// Get popular podcasts in category
	response, err := h.usecase.GetPopularInCategory(ctx, modelReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get popular podcasts in category: %v", err)
	}

	// Convert to gRPC response
	return convertToGRPCResponse(response), nil
}

// Helper function to convert model response to gRPC response
func convertToGRPCResponse(response *models.RecommendationResponse) *pb.GetRecommendationsResponse {
	var items []*pb.RecommendedItem
	for _, item := range response.Items {
		itemType := pb.RecommendedItem_PODCAST
		if item.Type == "episode" {
			itemType = pb.RecommendedItem_EPISODE
		}

		grpcItem := &pb.RecommendedItem{
			Id:          item.ID.String(),
			Type:        itemType,
			Title:       item.Title,
			Description: item.Description,
			ImageUrl:    item.ImageURL,
			Score:       float32(item.Score),
		}

		if item.PodcastID != uuid.Nil {
			grpcItem.PodcastId = item.PodcastID.String()
		}

		if item.PodcastTitle != "" {
			grpcItem.PodcastTitle = item.PodcastTitle
		}

		items = append(items, grpcItem)
	}

	return &pb.GetRecommendationsResponse{
		Items: items,
	}
}