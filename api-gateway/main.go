package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	orderpb "order-service/proto/order"
	restaurantpb "restaurant-service/proto"
	userpb "user-service/proto/user"
)

var (
	jwtSecret = []byte(getEnv("JWT_SECRET", "super-secret-key-369"))
)

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))

	limiter := rate.NewLimiter(10, 20)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	userConn, err := grpc.Dial(getEnv("USER_SERVICE_URL", "localhost:50051"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to user service: %v", err)
	}
	userClient := userpb.NewUserServiceClient(userConn)

	orderConn, err := grpc.Dial(getEnv("ORDER_SERVICE_URL", "localhost:50052"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to order service: %v", err)
	}
	orderClient := orderpb.NewOrderServiceClient(orderConn)

	restConn, err := grpc.Dial(getEnv("RESTAURANT_SERVICE_URL", "localhost:50053"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to restaurant service: %v", err)
	}
	restClient := restaurantpb.NewRestaurantServiceClient(restConn)

	r.Post("/auth/register", func(w http.ResponseWriter, r *http.Request) {
		var req userpb.RegisterRequest
		json.NewDecoder(r.Body).Decode(&req)
		res, err := userClient.Register(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		json.NewEncoder(w).Encode(res)
	})

	r.Post("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		var req userpb.LoginRequest
		json.NewDecoder(r.Body).Decode(&req)
		res, err := userClient.Login(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}
		json.NewEncoder(w).Encode(res)
	})

	r.Get("/restaurants", func(w http.ResponseWriter, req *http.Request) {
		res, err := restClient.ListRestaurants(req.Context(), &restaurantpb.ListRestaurantsRequest{})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(res)
	})

	r.Get("/items/search", func(w http.ResponseWriter, req *http.Request) {
		q := req.URL.Query().Get("query")
		res, err := restClient.SearchItems(req.Context(), &restaurantpb.SearchItemsRequest{Query: q})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(res)
	})

	r.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				authHeader := req.Header.Get("Authorization")
				if authHeader == "" {
					http.Error(w, "Unauthorized", 401)
					return
				}
				tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
				token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
					return jwtSecret, nil
				})
				if err != nil || !token.Valid {
					http.Error(w, "Unauthorized", 401)
					return
				}
				claims, ok := token.Claims.(jwt.MapClaims)
				if !ok {
					http.Error(w, "Unauthorized", 401)
					return
				}
				ctx := context.WithValue(req.Context(), "user_id", claims["user_id"])
				ctx = context.WithValue(ctx, "is_student", claims["is_student"])
				next.ServeHTTP(w, req.WithContext(ctx))
			})
		})

		r.Get("/profile", func(w http.ResponseWriter, req *http.Request) {
			userID := req.Context().Value("user_id").(string)
			res, err := userClient.GetProfile(req.Context(), &userpb.GetProfileRequest{UserId: userID})
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			json.NewEncoder(w).Encode(res)
		})

		r.Post("/orders", func(w http.ResponseWriter, req *http.Request) {
			userID := req.Context().Value("user_id").(string)

			isStudent := false
			if val, ok := req.Context().Value("is_student").(bool); ok {
				isStudent = val
			}

			var createReq orderpb.CreateOrderRequest
			json.NewDecoder(req.Body).Decode(&createReq)
			createReq.UserId = userID
			createReq.IsStudent = isStudent

			res, err := orderClient.CreateOrder(req.Context(), &createReq)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			json.NewEncoder(w).Encode(res)
		})

		r.Get("/orders", func(w http.ResponseWriter, req *http.Request) {
			userID := req.Context().Value("user_id").(string)
			res, err := orderClient.ListUserOrders(req.Context(), &orderpb.ListUserOrdersRequest{UserId: userID, Page: 1, Limit: 100})
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			json.NewEncoder(w).Encode(res)
		})
	})

	fmt.Println("API Gateway running on :8080")
	http.ListenAndServe(":8080", r)
}
