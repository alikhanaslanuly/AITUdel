package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"user-service/internal/config"
	grpcdelivery "user-service/internal/delivery/grpc"
	"user-service/internal/repository"
	"user-service/internal/usecase"
	"user-service/pkg/cache"
	"user-service/pkg/database"
	"user-service/pkg/email"
	jwtpkg "user-service/pkg/jwt"
	"user-service/pkg/messaging"
	"user-service/pkg/observability"
	pb "user-service/proto/user"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx := context.Background()

	shutdownObs, err := observability.Init(ctx, "user-service", getEnv("METRICS_PORT", "9102"))
	if err != nil {
		log.Fatalf("observability: %v", err)
	}
	defer func() {
		_ = shutdownObs(context.Background())
	}()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.NewPostgres(cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer func(db *sql.DB) {
		if err := db.Close(); err != nil {
			log.Printf("db close error: %v", err)
		}
	}(db)

	if err := database.RunMigrationsWithDSN("postgres://" + cfg.DBUser + ":" + cfg.DBPassword +
		"@" + cfg.DBHost + ":" + cfg.DBPort + "/" + cfg.DBName + "?sslmode=disable"); err != nil {
		log.Fatalf("migrations: %v", err)
	}
	log.Println("migrations done")

	redisClient, err := cache.NewRedis(cfg.RedisAddr())
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer func(redisClient *cache.RedisClient) {
		if err := redisClient.Close(); err != nil {
			log.Printf("redis close error: %v", err)
		}
	}(redisClient)

	natsClient, err := messaging.NewNats(cfg.NatsURL)
	if err != nil {
		log.Fatalf("nats: %v", err)
	}
	defer natsClient.Close()

	jwtManager, err := jwtpkg.NewManager(cfg.JWTSecret, cfg.JWTAccessExpMin, cfg.JWTRefreshExpHours)
	if err != nil {
		log.Fatalf("jwt: %v", err)
	}

	smtpSender, err := email.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPFrom)
	if err != nil {
		log.Fatalf("smtp: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	courierRepo := repository.NewCourierRepository(db)
	notifRepo := repository.NewNotifRepository(db)

	authUC := usecase.NewAuthUsecase(userRepo, tokenRepo, courierRepo, notifRepo, redisClient, jwtManager, natsClient)
	notifUC := usecase.NewNotifUsecase(userRepo, notifRepo, smtpSender)
	courierUC := usecase.NewCourierUsecase(courierRepo)

	grpcdelivery.SubscribeNATS(natsClient, notifUC)

	handler := grpcdelivery.NewUserHandler(authUC, notifUC, courierUC)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	server := googlegrpc.NewServer(
		googlegrpc.StatsHandler(otelgrpc.NewServerHandler()),
		googlegrpc.ChainUnaryInterceptor(
			observability.GRPCMetricsUnaryInterceptor("user-service"),
		),
	)

	pb.RegisterUserServiceServer(server, handler)
	reflection.Register(server)

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop

		log.Println("shutting down user-service...")
		server.GracefulStop()
	}()

	log.Printf("user-service gRPC running on :%s", cfg.GRPCPort)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
