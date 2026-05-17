package grpc

import (
	"context"

	"restaurant-service/internal/domain"
	"restaurant-service/internal/usecase"
	pb "restaurant-service/proto"
)

type Server struct {
	pb.UnimplementedRestaurantServiceServer
	usecase *usecase.RestaurantUsecase
}

func NewServer(
	usecase *usecase.RestaurantUsecase,
) *Server {

	return &Server{
		usecase: usecase,
	}
}

func (s *Server) ListRestaurants(
	ctx context.Context,
	req *pb.ListRestaurantsRequest,
) (*pb.ListRestaurantsResponse, error) {

	restaurants, err := s.usecase.ListRestaurants(ctx)
	if err != nil {
		return nil, err
	}

	var response []*pb.Restaurant

	for _, r := range restaurants {

		response = append(response, &pb.Restaurant{
			Id:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Rating:      r.Rating,
		})
	}

	return &pb.ListRestaurantsResponse{
		Restaurants: response,
	}, nil
}

func (s *Server) GetRestaurant(
	ctx context.Context,
	req *pb.GetRestaurantRequest,
) (*pb.RestaurantResponse, error) {

	r, err := s.usecase.GetRestaurant(
		ctx,
		req.Id,
	)

	if err != nil {
		return nil, err
	}

	return &pb.RestaurantResponse{
		Restaurant: &pb.Restaurant{
			Id:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Rating:      r.Rating,
		},
	}, nil
}

func (s *Server) SearchItems(
	ctx context.Context,
	req *pb.SearchItemsRequest,
) (*pb.SearchItemsResponse, error) {

	items, err := s.usecase.SearchItems(
		ctx,
		req.Query,
	)

	if err != nil {
		return nil, err
	}

	var response []*pb.MenuItem

	for _, item := range items {

		response = append(response, &pb.MenuItem{
			Id:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			Price:       item.Price,
			Stock:       int32(item.Stock),
		})
	}

	return &pb.SearchItemsResponse{
		Items: response,
	}, nil
}

func (s *Server) GetItem(
	ctx context.Context,
	req *pb.GetItemRequest,
) (*pb.ItemResponse, error) {

	item, err := s.usecase.GetItem(
		ctx,
		req.Id,
	)

	if err != nil {
		return nil, err
	}

	return &pb.ItemResponse{
		Item: &pb.MenuItem{
			Id:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			Price:       item.Price,
			Stock:       int32(item.Stock),
		},
	}, nil
}

func (s *Server) RateRestaurant(
	ctx context.Context,
	req *pb.RateRestaurantRequest,
) (*pb.RateRestaurantResponse, error) {

	review := domain.Review{
		RestaurantID: req.RestaurantId,
		UserID:       req.UserId,
		Rating:       int(req.Rating),
		Comment:      req.Comment,
	}

	err := s.usecase.RateRestaurant(
		ctx,
		review,
	)

	if err != nil {
		return nil, err
	}

	return &pb.RateRestaurantResponse{
		Message: "Review added successfully",
	}, nil
}
