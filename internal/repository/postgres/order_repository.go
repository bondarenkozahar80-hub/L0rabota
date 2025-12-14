package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"order-service0/internal/domain/entities"

	"github.com/pkg/errors"
)

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *orderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *entities.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	orderQuery := `INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, 
	                  customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) 
	                  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err = tx.ExecContext(ctx, orderQuery,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.ShardKey, order.SMID, order.DateCreated, order.OOFShard)
	if err != nil {
		return errors.Wrap(err, "failed to insert order")
	}

	deliveryQuery := `INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email) 
	                  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err = tx.ExecContext(ctx, deliveryQuery,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return errors.Wrap(err, "failed to insert delivery")
	}

	paymentQuery := `INSERT INTO payments (order_uid, transaction, request_id, currency, provider, 
	                  amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) 
	                  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err = tx.ExecContext(ctx, paymentQuery,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDT, order.Payment.Bank,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return errors.Wrap(err, "failed to insert payment")
	}

	itemQuery := `INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, 
	                total_price, nm_id, brand, status) 
	                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx, itemQuery,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NMID, item.Brand, item.Status)
		if err != nil {
			return errors.Wrap(err, "failed to insert item")
		}
	}

	return tx.Commit()
}

func (r *orderRepository) GetByUID(ctx context.Context, orderUID string) (*entities.Order, error) {
	query := `
		SELECT o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, 
		       o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
		       d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
		       p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, 
		       p.bank, p.delivery_cost, p.goods_total, p.custom_fee
		FROM orders o
		LEFT JOIN deliveries d ON o.order_uid = d.order_uid
		LEFT JOIN payments p ON o.order_uid = p.order_uid
		WHERE o.order_uid = $1`

	var order entities.Order
	var delivery entities.Delivery
	var payment entities.Payment

	err := r.db.QueryRowContext(ctx, query, orderUID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
		&order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SMID, &order.DateCreated, &order.OOFShard,
		&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address, &delivery.Region, &delivery.Email,
		&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider, &payment.Amount,
		&payment.PaymentDT, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, errors.Wrap(err, "failed to get order")
	}

	order.Delivery = delivery
	order.Payment = payment

	items, err := r.getItemsByOrderUID(ctx, orderUID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get items")
	}
	order.Items = items

	return &order, nil
}

func (r *orderRepository) getItemsByOrderUID(ctx context.Context, orderUID string) ([]entities.Item, error) {
	query := `SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status 
	          FROM items WHERE order_uid = $1`
	rows, err := r.db.QueryContext(ctx, query, orderUID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query items")
	}
	defer rows.Close()

	var items []entities.Item
	for rows.Next() {
		var item entities.Item
		if err := rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name, &item.Sale,
			&item.Size, &item.TotalPrice, &item.NMID, &item.Brand, &item.Status,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan item")
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *orderRepository) GetAll(ctx context.Context) ([]*entities.Order, error) {
	query := `SELECT order_uid FROM orders`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get order UIDs")
	}
	defer rows.Close()

	var orders []*entities.Order
	for rows.Next() {
		var orderUID string
		if err := rows.Scan(&orderUID); err != nil {
			return nil, errors.Wrap(err, "failed to scan order UID")
		}
		order, err := r.GetByUID(ctx, orderUID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get order %s", orderUID)
		}
		orders = append(orders, order)
	}

	return orders, nil
}
