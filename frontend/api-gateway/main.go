package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

type rawClient struct {
	conn *grpc.ClientConn
}

func (c *rawClient) call(method string, req, resp any) error {
	cx, cancel := ctx()
	defer cancel()
	return c.conn.Invoke(cx, method, req, resp)
}

// User Service
type RegisterReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
}

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateProfileReq struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Phone  string `json:"phone"`
}

// Order Service
type OrderItemReq struct {
	ItemID   string  `json:"item_id"`
	Name     string  `json:"name"`
	Quantity int32   `json:"quantity"`
	Price    float64 `json:"price"`
}

type CreateOrderReq struct {
	UserID          string         `json:"user_id"`
	RestaurantID    string         `json:"restaurant_id"`
	DeliveryAddress string         `json:"delivery_address"`
	Items           []OrderItemReq `json:"items"`
	PromoCode       string         `json:"promo_code"`
	IsStudent       bool           `json:"is_student"`
}

type AddToCartReq struct {
	UserID string       `json:"user_id"`
	Item   OrderItemReq `json:"item"`
}

type ApplyPromoReq struct {
	OrderID    string  `json:"order_id"`
	UserID     string  `json:"user_id"`
	PromoCode  string  `json:"promo_code"`
	IsStudent  bool    `json:"is_student"`
	OrderTotal float64 `json:"order_total"`
}

type CancelOrderReq struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
}

// Restaurant Service
type RateReq struct {
	RestaurantID int64  `json:"restaurant_id"`
	UserID       int64  `json:"user_id"`
	Rating       int32  `json:"rating"`
	Comment      string `json:"comment"`
}

// ─── Хендлеры User Service ──────────────────────────────────────

func handleRegister(userConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterReq
		if err := decodeBody(r, &req); err != nil {
			writeError(w, 400, "bad request")
			return
		}
		if req.Role == "" {
			req.Role = "user"
		}

		var resp map[string]any
		c := &rawClient{conn: userConn}
		if err := c.call("/user.UserService/Register", req, &resp); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleLogin(userConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginReq
		if err := decodeBody(r, &req); err != nil {
			writeError(w, 400, "bad request")
			return
		}
		var resp map[string]any
		c := &rawClient{conn: userConn}
		if err := c.call("/user.UserService/Login", req, &resp); err != nil {
			writeError(w, 401, "invalid credentials")
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleGetProfile(userConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		var resp map[string]any
		c := &rawClient{conn: userConn}
		if err := c.call("/user.UserService/GetProfile", map[string]string{"user_id": userID}, &resp); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleUpdateProfile(userConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req UpdateProfileReq
		if err := decodeBody(r, &req); err != nil {
			writeError(w, 400, "bad request")
			return
		}
		var resp map[string]any
		c := &rawClient{conn: userConn}
		if err := c.call("/user.UserService/UpdateProfile", req, &resp); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleLogout(userConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req map[string]string
		decodeBody(r, &req)
		var resp map[string]any
		c := &rawClient{conn: userConn}
		c.call("/user.UserService/Logout", req, &resp)
		writeJSON(w, 200, map[string]bool{"success": true})
	}
}

// ─── Хендлеры Restaurant Service ────────────────────────────────

func handleListRestaurants(restConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resp map[string]any
		c := &rawClient{conn: restConn}
		if err := c.call("/restaurant.RestaurantService/ListRestaurants", map[string]any{}, &resp); err != nil {
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
		var resp map[string]any
		c := &rawClient{conn: restConn}
		if err := c.call("/restaurant.RestaurantService/GetRestaurant", map[string]any{"id": id}, &resp); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleSearchItems(restConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		var resp map[string]any
		c := &rawClient{conn: restConn}
		if err := c.call("/restaurant.RestaurantService/SearchItems", map[string]string{"query": q}, &resp); err != nil {
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
		var resp map[string]any
		c := &rawClient{conn: restConn}
		if err := c.call("/restaurant.RestaurantService/GetItem", map[string]any{"id": id}, &resp); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleRateRestaurant(restConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RateReq
		if err := decodeBody(r, &req); err != nil {
			writeError(w, 400, "bad request")
			return
		}
		var resp map[string]any
		c := &rawClient{conn: restConn}
		if err := c.call("/restaurant.RestaurantService/RateRestaurant", req, &resp); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

// ─── Хендлеры Order Service ─────────────────────────────────────

func handleCreateOrder(orderConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateOrderReq
		if err := decodeBody(r, &req); err != nil {
			writeError(w, 400, "bad request")
			return
		}
		var resp map[string]any
		c := &rawClient{conn: orderConn}
		if err := c.call("/order.OrderService/CreateOrder", req, &resp); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleGetOrder(orderConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := r.URL.Query().Get("order_id")
		var resp map[string]any
		c := &rawClient{conn: orderConn}
		if err := c.call("/order.OrderService/GetOrder", map[string]string{"order_id": orderID}, &resp); err != nil {
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
		var resp map[string]any
		c := &rawClient{conn: orderConn}
		if err := c.call("/order.OrderService/ListUserOrders", map[string]any{
			"user_id": userID,
			"page":    int32(page),
			"limit":   int32(limit),
		}, &resp); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleCancelOrder(orderConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CancelOrderReq
		if err := decodeBody(r, &req); err != nil {
			writeError(w, 400, "bad request")
			return
		}
		var resp map[string]any
		c := &rawClient{conn: orderConn}
		if err := c.call("/order.OrderService/CancelOrder", req, &resp); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleApplyPromo(orderConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ApplyPromoReq
		if err := decodeBody(r, &req); err != nil {
			writeError(w, 400, "bad request")
			return
		}
		var resp map[string]any
		c := &rawClient{conn: orderConn}
		if err := c.call("/order.OrderService/ApplyPromo", req, &resp); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleAddToCart(orderConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AddToCartReq
		if err := decodeBody(r, &req); err != nil {
			writeError(w, 400, "bad request")
			return
		}
		var resp map[string]any
		c := &rawClient{conn: orderConn}
		if err := c.call("/order.OrderService/AddToCart", req, &resp); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleGetCart(orderConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		var resp map[string]any
		c := &rawClient{conn: orderConn}
		if err := c.call("/order.OrderService/GetCart", map[string]string{"user_id": userID}, &resp); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

func handleClearCart(orderConn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req map[string]string
		decodeBody(r, &req)
		var resp map[string]any
		c := &rawClient{conn: orderConn}
		if err := c.call("/order.OrderService/ClearCart", req, &resp); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

// ─── Роутер ─────────────────────────────────────────────────────

func main() {
	userConn := connectGRPC("localhost:50052")
	orderConn := connectGRPC("localhost:50051")
	restConn := connectGRPC("localhost:50053")
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
