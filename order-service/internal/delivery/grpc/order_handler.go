package grpc

import (
	"context"
	"order-service/internal/domain"
	"order-service/internal/usecase"
	pb "order-service/proto/order"

	"github.com/google/uuid"
)

type OrderHandler struct {
	pb.UnimplementedOrderServiceServer
	orderUC *usecase.OrderUsecase
	cartUC  *usecase.CartUsecase
}

func NewOrderHandler(orderUC *usecase.OrderUsecase, cartUC *usecase.CartUsecase) *OrderHandler {
	return &OrderHandler{
		orderUC: orderUC,
		cartUC:  cartUC,
	}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	items := make([]domain.OrderItem, len(req.Items))
	for i, pbItem := range req.Items {
		items[i] = domain.OrderItem{
			ItemID:   pbItem.ItemId,
			Name:     pbItem.Name,
			Quantity: int(pbItem.Quantity),
			Price:    pbItem.Price,
		}
	}

	order := &domain.Order{
		UserID:          req.UserId,
		RestaurantID:    req.RestaurantId,
		DeliveryAddress: req.DeliveryAddress,
		Items:           items,
	}

	created, err := h.orderUC.CreateOrder(ctx, order, req.PromoCode, req.IsStudent)
	if err != nil {
		return nil, err
	}

	return &pb.CreateOrderResponse{Order: domainToProto(created)}, nil
}

func (h *OrderHandler) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	order, err := h.orderUC.GetOrder(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}
	return &pb.GetOrderResponse{Order: domainToProto(order)}, nil
}

func (h *OrderHandler) ListUserOrders(ctx context.Context, req *pb.ListUserOrdersRequest) (*pb.ListUserOrdersResponse, error) {
	orders, total, err := h.orderUC.ListUserOrders(ctx, req.UserId, int(req.Page), int(req.Limit))
	if err != nil {
		return nil, err
	}

	pbOrders := make([]*pb.Order, len(orders))
	for i, o := range orders {
		pbOrders[i] = domainToProto(o)
	}

	return &pb.ListUserOrdersResponse{Orders: pbOrders, Total: int32(total)}, nil
}

func (h *OrderHandler) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	if err := h.orderUC.CancelOrder(ctx, req.OrderId, req.UserId); err != nil {
		return &pb.CancelOrderResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb.CancelOrderResponse{Success: true, Message: "order cancelled"}, nil
}

func (h *OrderHandler) ApplyPromo(ctx context.Context, req *pb.ApplyPromoRequest) (*pb.ApplyPromoResponse, error) {
	discount, newTotal, err := h.orderUC.ApplyPromo(ctx, req.OrderId, req.UserId, req.PromoCode, req.IsStudent, req.OrderTotal)
	if err != nil {
		return nil, err
	}
	return &pb.ApplyPromoResponse{
		DiscountAmount: discount,
		NewTotal:       newTotal,
		Message:        "promo applied",
	}, nil
}

func (h *OrderHandler) AddToCart(ctx context.Context, req *pb.AddToCartRequest) (*pb.AddToCartResponse, error) {
	item := domain.CartItem{
		ID:       uuid.New().String(),
		UserID:   req.UserId,
		ItemID:   req.Item.ItemId,
		Name:     req.Item.Name,
		Quantity: int(req.Item.Quantity),
		Price:    req.Item.Price,
	}

	if err := h.cartUC.AddItem(ctx, req.UserId, item); err != nil {
		return &pb.AddToCartResponse{Success: false}, nil
	}
	return &pb.AddToCartResponse{Success: true}, nil
}

func (h *OrderHandler) GetCart(ctx context.Context, req *pb.GetCartRequest) (*pb.GetCartResponse, error) {
	items, total, err := h.cartUC.GetCart(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	pbItems := make([]*pb.CartItem, len(items))
	for i, item := range items {
		pbItems[i] = &pb.CartItem{
			ItemId:   item.ItemID,
			Name:     item.Name,
			Quantity: int32(item.Quantity),
			Price:    item.Price,
		}
	}

	return &pb.GetCartResponse{Items: pbItems, Total: total}, nil
}

func (h *OrderHandler) ClearCart(ctx context.Context, req *pb.ClearCartRequest) (*pb.ClearCartResponse, error) {
	if err := h.cartUC.ClearCart(ctx, req.UserId); err != nil {
		return &pb.ClearCartResponse{Success: false}, nil
	}
	return &pb.ClearCartResponse{Success: true}, nil
}

func domainToProto(o *domain.Order) *pb.Order {
	pbItems := make([]*pb.OrderItem, len(o.Items))
	for i, item := range o.Items {
		pbItems[i] = &pb.OrderItem{
			ItemId:   item.ItemID,
			Name:     item.Name,
			Quantity: int32(item.Quantity),
			Price:    item.Price,
		}
	}

	return &pb.Order{
		Id:              o.ID,
		UserId:          o.UserID,
		Status:          o.Status,
		TotalPrice:      o.TotalPrice,
		PromoCode:       o.PromoCode,
		DiscountAmount:  o.DiscountAmount,
		DeliveryAddress: o.DeliveryAddress,
		RestaurantId:    o.RestaurantID,
		Items:           pbItems,
		CreatedAt:       o.CreatedAt.String(),
	}
}
