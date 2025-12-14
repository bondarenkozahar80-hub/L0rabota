package usecase

import (
	"context"
	"encoding/json"
	"order-service0/internal/domain/entities"
	"order-service0/internal/pkg/validator"

	"github.com/pkg/errors"
)

type orderUseCase struct {
	orderRepo OrderRepository
	cache     Cache
	validator *validator.CustomValidator
}

func NewOrderUseCase(orderRepo OrderRepository, cache Cache) OrderUseCase {
	return &orderUseCase{
		orderRepo: orderRepo,
		cache:     cache,
		validator: validator.NewValidator(),
	}
}

func (uc *orderUseCase) CreateOrder(ctx context.Context, order *entities.Order) error {
	// Валидация данных
	if err := uc.validator.ValidateStruct(order); err != nil {
		return errors.Wrap(err, "order validation failed")
	}

	// Сохранение в базу данных
	if err := uc.orderRepo.Create(ctx, order); err != nil {
		return errors.Wrap(err, "failed to save order to database")
	}

	// Кэширование заказа
	uc.cache.Set(order.OrderUID, order)
	return nil
}

func (uc *orderUseCase) GetOrderByUID(ctx context.Context, orderUID string) (*entities.Order, error) {
	// Поиск в кэше
	if order, exists := uc.cache.Get(orderUID); exists {
		return order, nil
	}

	// Поиск в базе данных
	order, err := uc.orderRepo.GetByUID(ctx, orderUID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get order from database")
	}

	// Сохранение в кэш для будущих запросов
	uc.cache.Set(orderUID, order)
	return order, nil
}

func (uc *orderUseCase) ProcessOrderMessage(ctx context.Context, message []byte) error {
	var order entities.Order
	if err := json.Unmarshal(message, &order); err != nil {
		return errors.Wrap(err, "failed to unmarshal order message")
	}

	return uc.CreateOrder(ctx, &order)
}
