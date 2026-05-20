package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	grpcHandler "restaurant-service/internal/delivery/grpc"
	"restaurant-service/internal/repository"
	"restaurant-service/internal/usecase"

	"restaurant-service/pkg/cache"
	"restaurant-service/pkg/database"
	"restaurant-service/pkg/messaging"

	pb "restaurant-service/proto"
)

func main() {

	db, err := database.NewPostgresConnection()
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	redisClient := cache.NewRedisClient()

	natsConn := messaging.NewNATSConnection()

	repo := repository.NewPostgresRepository(db)

	restaurantUsecase := usecase.NewRestaurantUsecase(
		repo,
		redisClient,
	)

	stockUsecase := usecase.NewStockUsecase(db)

	messaging.SubscribeOrderCreated(
		natsConn,
		stockUsecase,
	)

	server := grpcHandler.NewServer(
		restaurantUsecase,
	)

	lis, err := net.Listen(
		"tcp",
		":50053",
	)

	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterRestaurantServiceServer(
		grpcServer,
		server,
	)

	log.Println(
		"Restaurant Service running on :50053",
	)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
