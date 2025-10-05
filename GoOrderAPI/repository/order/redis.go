package order

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hpharis258/orders-api/model"
	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	Client *redis.Client
}

func orderIDKey(orderID uint64) string {
	return fmt.Sprintf("order:%d", orderID)
}

func (r *RedisRepo) Insert(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}
	key := orderIDKey(order.OrderID)
	txn := r.Client.TxPipeline()
	res := txn.SetNX(ctx, key, data, 0)
	if res.Err() != nil {
		txn.Discard()
		return fmt.Errorf("failed to insert order into Redis: %w", res.Err())
	}
	if err := txn.SAdd(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to add order ID to orders set: %w", err)
	}
	_, err = txn.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute Redis transaction: %w", err)
	}
	return nil
}

func (r *RedisRepo) GetByID(ctx context.Context, orderID uint64) (*model.Order, error) {
	key := orderIDKey(orderID)
	res := r.Client.Get(ctx, key)
	if res.Err() != nil {
		if res.Err() == redis.Nil {
			return nil, nil // Order not found
		}
		return nil, fmt.Errorf("failed to get order from Redis: %w", res.Err())
	}
	var order model.Order
	if err := json.Unmarshal([]byte(res.Val()), &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order data: %w", err)
	}
	return &order, nil
}

func (r *RedisRepo) DeleteById(ctx context.Context, id uint64) error {
	key := orderIDKey(id)
	txn := r.Client.TxPipeline()
	res := txn.Del(ctx, key)
	if res.Err() != nil {
		txn.Discard()
		return fmt.Errorf("failed to delete order from Redis: %w", res.Err())
	}
	if err := txn.SRem(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to remove order ID from orders set: %w", err)
	}
	_, err := txn.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute Redis transaction: %w", err)
	}
	return nil
}
func (r *RedisRepo) Update(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}
	key := orderIDKey(order.OrderID)
	err = r.Client.SetXX(ctx, key, string(data), 0).Err()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("order with ID %d does not exist", order.OrderID)
		}
		return fmt.Errorf("failed to update order in Redis: %w", err)
	}
	return nil
}

type FindAllPage struct {
	Size   uint
	Offset uint64
}

type FindResult struct {
	Orders []model.Order
	Cursor uint64
}

func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	res := r.Client.SScan(ctx, "orders", page.Offset, "*", int64(page.Size))
	if res.Err() != nil {
		return FindResult{}, fmt.Errorf("failed to scan orders set: %w", res.Err())
	}
	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get scan result: %w", err)
	}

	if len(keys) == 0 {
		return FindResult{
			Orders: []model.Order{},
			Cursor: cursor,
		}, nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get orders from Redis: %w", err)
	}

	orders := make([]model.Order, 0, len(xs))
	for _, x := range xs {
		if x == nil {
			continue // Skip nil values
		}
		var order model.Order
		if err := json.Unmarshal([]byte(x.(string)), &order); err != nil {
			return FindResult{}, fmt.Errorf("failed to unmarshal order data: %w", err)
		}
		orders = append(orders, order)
	}

	return FindResult{
		Orders: orders,
		Cursor: cursor,
	}, nil
}
