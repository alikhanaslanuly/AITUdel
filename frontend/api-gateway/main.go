package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	orderpb "api-gateway/proto/order"
	restpb "api-gateway/proto/restaurant"
	userpb "api-gateway/proto/user"
)

type GatewayClients struct {
	userConn       *grpc.ClientConn
	orderConn      *grpc.ClientConn
	restaurantConn *grpc.ClientConn
}

func connectGRPC(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("grpc dial %s: %v", addr, err)
	}
	return conn
}

func writeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func decodeBody(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

func ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

func handleRegister(userConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Name     string `json:"name"`
			Phone    string `json:"phone"`
			Role     string `json:"role"`
		}
		if err := decodeBody(r, &body); err != nil {
			writeError(w, 400, "bad request")
			return
		}
		if body.Role == "" {
			body.Role = "user"
		}

		c := userpb.NewUserServiceClient(userConn)
		cx, cancel := ctx()
		defer cancel()

		resp, err := c.Register(cx, &userpb.RegisterRequest{
			Email:    body.Email,
			Password: body.Password,
			Name:     body.Name,
			Phone:    body.Phone,
			Role:     body.Role,
		})
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleLogin(userConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := decodeBody(r, &body); err != nil {
			writeError(w, 400, "bad request")
			return
		}

		c := userpb.NewUserServiceClient(userConn)
		cx, cancel := ctx()
		defer cancel()

		resp, err := c.Login(cx, &userpb.LoginRequest{
			Email:    body.Email,
			Password: body.Password,
		})
		if err != nil {
			writeError(w, 401, "invalid credentials")
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleGetProfile(userConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")

		c := userpb.NewUserServiceClient(userConn)
		cx, cancel := ctx()
		defer cancel()

		resp, err := c.GetProfile(cx, &userpb.GetProfileRequest{UserId: userID})
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleUpdateProfile(userConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			UserID string `json:"user_id"`
			Name   string `json:"name"`
			Phone  string `json:"phone"`
		}
		if err := decodeBody(r, &body); err != nil {
			writeError(w, 400, "bad request")
			return
		}

		c := userpb.NewUserServiceClient(userConn)
		cx, cancel := ctx()
		defer cancel()

		resp, err := c.UpdateProfile(cx, &userpb.UpdateProfileRequest{
			UserId: body.UserID,
			Name:   body.Name,
			Phone:  body.Phone,
		})
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleLogout(userConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			UserID       string `json:"user_id"`
			RefreshToken string `json:"refresh_token"`
		}
		decodeBody(r, &body)

		c := userpb.NewUserServiceClient(userConn)
		cx, cancel := ctx()
		defer cancel()

		c.Logout(cx, &userpb.LogoutRequest{ //nolint
			RefreshToken: body.RefreshToken,
		})
		writeJSON(w, 200, map[string]bool{"success": true})
	}
}

func handleListRestaurants(restConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := restpb.NewRestaurantServiceClient(restConn)
		cx, cancel := ctx()
		defer cancel()

		resp, err := c.ListRestaurants(cx, &restpb.ListRestaurantsRequest{})
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleGetRestaurant(restConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		id, _ := strconv.ParseInt(idStr, 10, 64)

		c := restpb.NewRestaurantServiceClient(restConn)
		cx, cancel := ctx()
		defer cancel()

		resp, err := c.GetRestaurant(cx, &restpb.GetRestaurantRequest{Id: id})
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleSearchItems(restConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")

		c := restpb.NewRestaurantServiceClient(restConn)
		cx, cancel := ctx()
		defer cancel()

		resp, err := c.SearchItems(cx, &restpb.SearchItemsRequest{Query: q})
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleGetItem(restConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		id, _ := strconv.ParseInt(idStr, 10, 64)

		c := restpb.NewRestaurantServiceClient(restConn)
		cx, cancel := ctx()
		defer cancel()

		resp, err := c.GetItem(cx, &restpb.GetItemRequest{Id: id})
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleRateRestaurant(restConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			RestaurantID int64  `json:"restaurant_id"`
			UserID       int64  `json:"user_id"`
			Rating       int32  `json:"rating"`
			Comment      string `json:"comment"`
		}
		if err := decodeBody(r, &body); err != nil {
			writeError(w, 400, "bad request")
			return
		}

		c := restpb.NewRestaurantServiceClient(restConn)
		cx, cancel := ctx()
		defer cancel()

		resp, err := c.RateRestaurant(cx, &restpb.RateRestaurantRequest{
			RestaurantId: body.RestaurantID,
			UserId:       body.UserID,
			Rating:       body.Rating,
			Comment:      body.Comment,
		})
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleCreateOrder(orderConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			UserID          string `json:"user_id"`
			RestaurantID    string `json:"restaurant_id"`
			DeliveryAddress string `json:"delivery_address"`
			PromoCode       string `json:"promo_code"`
			IsStudent       bool   `json:"is_student"`
			Items           []struct {
				ItemID   string  `json:"item_id"`
				Name     string  `json:"name"`
				Quantity int32   `json:"quantity"`
				Price    float64 `json:"price"`
			} `json:"items"`
		}
		if err := decodeBody(r, &body); err != nil {
			writeError(w, 400, "bad request")
			return
		}

		items := make([]*orderpb.OrderItem, len(body.Items))
		for i, it := range body.Items {
			items[i] = &orderpb.OrderItem{
				ItemId:   it.ItemID,
				Name:     it.Name,
				Quantity: it.Quantity,
				Price:    it.Price,
			}
		}

		c := orderpb.NewOrderServiceClient(orderConn)
		cx, cancel := ctx()
		defer cancel()

		resp, err := c.CreateOrder(cx, &orderpb.CreateOrderRequest{
			UserId:          body.UserID,
			RestaurantId:    body.RestaurantID,
			DeliveryAddress: body.DeliveryAddress,
			PromoCode:       body.PromoCode,
			IsStudent:       body.IsStudent,
			Items:           items,
		})
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleGetOrder(orderConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := r.URL.Query().Get("order_id")

		c := orderpb.NewOrderServiceClient(orderConn)
		cx, cancel := ctx()
		defer cancel()

		resp, err := c.GetOrder(cx, &orderpb.GetOrderRequest{OrderId: orderID})
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleListUserOrders(orderConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if page == 0 {
			page = 1
		}
		if limit == 0 {
			limit = 10
		}

		c := orderpb.NewOrderServiceClient(orderConn)
		cx, cancel := ctx()
		defer cancel()

		resp, err := c.ListUserOrders(cx, &orderpb.ListUserOrdersRequest{
			UserId: userID,
			Page:   int32(page),
			Limit:  int32(limit),
		})
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleCancelOrder(orderConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			OrderID string `json:"order_id"`
			UserID  string `json:"user_id"`
		}
		if err := decodeBody(r, &body); err != nil {
			writeError(w, 400, "bad request")
			return
		}

		c := orderpb.NewOrderServiceClient(orderConn)
		cx, cancel := ctx()
		defer cancel()

		resp, err := c.CancelOrder(cx, &orderpb.CancelOrderRequest{
			OrderId: body.OrderID,
			UserId:  body.UserID,
		})
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleApplyPromo(orderConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			OrderID    string  `json:"order_id"`
			UserID     string  `json:"user_id"`
			PromoCode  string  `json:"promo_code"`
			IsStudent  bool    `json:"is_student"`
			OrderTotal float64 `json:"order_total"`
		}
		if err := decodeBody(r, &body); err != nil {
			writeError(w, 400, "bad request")
			return
		}

		c := orderpb.NewOrderServiceClient(orderConn)
		cx, cancel := ctx()
		defer cancel()

		resp, err := c.ApplyPromo(cx, &orderpb.ApplyPromoRequest{
			OrderId:    body.OrderID,
			UserId:     body.UserID,
			PromoCode:  body.PromoCode,
			IsStudent:  body.IsStudent,
			OrderTotal: body.OrderTotal,
		})
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleAddToCart(orderConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			UserID string `json:"user_id"`
			Item   struct {
				ItemID   string  `json:"item_id"`
				Name     string  `json:"name"`
				Quantity int32   `json:"quantity"`
				Price    float64 `json:"price"`
			} `json:"item"`
		}
		if err := decodeBody(r, &body); err != nil {
			writeError(w, 400, "bad request")
			return
		}

		c := orderpb.NewOrderServiceClient(orderConn)
		cx, cancel := ctx()
		defer cancel()

		resp, err := c.AddToCart(cx, &orderpb.AddToCartRequest{
			UserId: body.UserID,
			Item: &orderpb.CartItem{
				ItemId:   body.Item.ItemID,
				Name:     body.Item.Name,
				Quantity: body.Item.Quantity,
				Price:    body.Item.Price,
			},
		})
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleGetCart(orderConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")

		c := orderpb.NewOrderServiceClient(orderConn)
		cx, cancel := ctx()
		defer cancel()

		resp, err := c.GetCart(cx, &orderpb.GetCartRequest{UserId: userID})
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleClearCart(orderConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			UserID string `json:"user_id"`
		}
		decodeBody(r, &body)

		c := orderpb.NewOrderServiceClient(orderConn)
		cx, cancel := ctx()
		defer cancel()

		resp, err := c.ClearCart(cx, &orderpb.ClearCartRequest{UserId: body.UserID})
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func main() {
	userConn := connectGRPC(getEnv("USER_SERVICE_ADDR", "localhost:50052"))
	orderConn := connectGRPC(getEnv("ORDER_SERVICE_ADDR", "localhost:50051"))
	restConn := connectGRPC(getEnv("RESTAURANT_SERVICE_ADDR", "localhost:50053"))
	defer userConn.Close()
	defer orderConn.Close()
	defer restConn.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/api/auth/register", handleRegister(userConn))
	mux.HandleFunc("/api/auth/login", handleLogin(userConn))
	mux.HandleFunc("/api/auth/logout", handleLogout(userConn))
	mux.HandleFunc("/api/user/profile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleGetProfile(userConn)(w, r)
		} else if r.Method == http.MethodPut {
			handleUpdateProfile(userConn)(w, r)
		}
	})

	mux.HandleFunc("/api/restaurants", handleListRestaurants(restConn))
	mux.HandleFunc("/api/restaurants/get", handleGetRestaurant(restConn))
	mux.HandleFunc("/api/restaurants/rate", handleRateRestaurant(restConn))
	mux.HandleFunc("/api/items/search", handleSearchItems(restConn))
	mux.HandleFunc("/api/items/get", handleGetItem(restConn))

	mux.HandleFunc("/api/orders/create", handleCreateOrder(orderConn))
	mux.HandleFunc("/api/orders/get", handleGetOrder(orderConn))
	mux.HandleFunc("/api/orders/list", handleListUserOrders(orderConn))
	mux.HandleFunc("/api/orders/cancel", handleCancelOrder(orderConn))
	mux.HandleFunc("/api/orders/promo", handleApplyPromo(orderConn))

	mux.HandleFunc("/api/cart/add", handleAddToCart(orderConn))
	mux.HandleFunc("/api/cart", handleGetCart(orderConn))
	mux.HandleFunc("/api/cart/clear", handleClearCart(orderConn))

	mux.Handle("/", http.FileServer(http.Dir("../public")))

	handler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	}).Handler(mux)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Println("API Gateway running on :8080")
	log.Println("Frontend: http://localhost:8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

	_ = strings.TrimSpace
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
