package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"effective-mobile-subscription-server/internal/domain/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5"

	"github.com/avast/retry-go/v5"
)

// RetriableError is a custom error that contains a positive duration for the next retry
type RetriableError struct {
	Err        error
	RetryAfter time.Duration
}

// Error returns error message and a Retry-After duration
func (e *RetriableError) Error() string {
	return fmt.Sprintf("%s (retry after %v)", e.Err.Error(), e.RetryAfter)
}

var _ error = (*RetriableError)(nil)

type IDatasource interface {
	AddSubscription(ctx context.Context, subscription *models.Subscription) error
	GetSubscriptionByUser(ctx context.Context, userID string) ([]models.Subscription, error)
	GetSubscriptionByName(ctx context.Context, subName string) ([]models.Subscription, error)
	GetSubscriptionByID(ctx context.Context, id int) ([]models.Subscription, error)
	UpdateSubscription(ctx context.Context, sub *models.Subscription) error
	DeleteSubscription(ctx context.Context, subID int) error
	GetSubscriptionsSum(ctx context.Context, userID, subName string, fromDate, toDate time.Time) (int, error)
	Stop(ctx context.Context) error
}
type Datasource struct {
	pool *pgxpool.Pool
	log  *logrus.Logger
}

func NewDatasource(ctx context.Context, dsn string, log *logrus.Logger) (IDatasource, error) {
	dbConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	var pool *pgxpool.Pool
	err = retry.New(
		retry.DelayType(func(n uint, err error, config retry.DelayContext) time.Duration {
			log.WithError(err).Error("Server fails with: ")
			if retriable, ok := err.(*RetriableError); ok {
				log.WithField("retry-after", retriable.RetryAfter.String()).Info("Server retry in: ")
				return retriable.RetryAfter
			}
			// apply a default exponential back off strategy
			return retry.BackOffDelay(n, err, config)
		}),
	).Do(
		func() error {
			var err error
			pool, err = pgxpool.NewWithConfig(ctx, dbConfig)
			if err != nil {
				return err
			}
			return nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create database connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.WithField("dsn", dsn).Info("database connection pool established")

	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		log.WithError(err).Error("migration init failed")
		return nil, err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.WithError(err).Error("migration failed")
		return nil, err
	}

	log.Info("database migration completed")

	return &Datasource{
		pool: pool,
		log:  log,
	}, nil
}

func (d Datasource) AddSubscription(ctx context.Context, subscription *models.Subscription) error {
	query := `INSERT INTO subscriptions (user_id, name, price, start_date, end_date)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (user_id, name, start_date)
	DO UPDATE SET price=EXCLUDED.price, end_date=EXCLUDED.end_date
	RETURNING id, (xmax = 0) AS inserted;`

	d.log.WithFields(logrus.Fields{
		"user_id": subscription.UserID,
		"name":    subscription.ServiceName,
	}).Debug("adding subscription")

	tx, err := d.pool.Begin(ctx)
	if err != nil {
		d.log.WithError(err).Error("failed to start transaction")
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	var inserted bool
	var id int
	if err := tx.QueryRow(ctx, query,
		subscription.UserID,
		subscription.ServiceName,
		subscription.Price,
		subscription.StartDate,
		subscription.EndDate,
	).Scan(&id, &inserted); err != nil {

		tx.Rollback(ctx)

		d.log.WithError(err).WithFields(logrus.Fields{
			"user_id": subscription.UserID,
			"name":    subscription.ServiceName,
		}).Error("failed to insert subscription")

		return fmt.Errorf("failed to add subscription: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		d.log.WithError(err).Error("failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	if !inserted {
		d.log.WithFields(logrus.Fields{
			"id": id,
		}).Info("subscription updated")
	} else {
		d.log.Info("subscription inserted")
	}

	return nil
}

func (d Datasource) GetSubscriptionByID(ctx context.Context, id int) ([]models.Subscription, error) {
	d.log.WithField("id", id).Debug("query subscriptions by id")

	query := `SELECT id, name, user_id, price, start_date, end_date
	          FROM subscriptions WHERE id = $1`

	return d.selectQueryWithParam(ctx, query, []any{id})
}

func (d Datasource) GetSubscriptionByUser(ctx context.Context, userID string) ([]models.Subscription, error) {
	d.log.WithField("user_id", userID).Debug("query subscriptions by user")

	query := `SELECT id, name, user_id, price, start_date, end_date
	          FROM subscriptions WHERE user_id = $1`

	return d.selectQueryWithParam(ctx, query, []any{userID})
}

func (d Datasource) GetSubscriptionByName(ctx context.Context, subName string) ([]models.Subscription, error) {
	d.log.WithField("subscription_name", subName).Debug("query subscriptions by name")

	query := `SELECT id, name, user_id, price, start_date, end_date
	          FROM subscriptions WHERE name = $1`

	return d.selectQueryWithParam(ctx, query, []any{subName})
}

func (d Datasource) selectQueryWithParam(
	ctx context.Context,
	query string,
	args []any,
) ([]models.Subscription, error) {

	rows, err := d.pool.Query(ctx, query, args...)
	if err != nil {
		d.log.WithError(err).Error("failed to execute query")
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	subs, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Subscription])
	if err != nil {
		d.log.WithError(err).Error("failed to scan rows")
		return nil, fmt.Errorf("failed to scan rows: %w", err)
	}

	return subs, nil
}

func (d Datasource) UpdateSubscription(ctx context.Context, sub *models.Subscription) error {
	query := `UPDATE subscriptions SET name=$1, price=$2, start_date=$3, end_date=$4, user_id=$5 WHERE id = $6;`

	d.log.WithFields(logrus.Fields{"subscription": sub}).Debug("updating subscription")

	tx, err := d.pool.Begin(ctx)
	if err != nil {
		d.log.WithError(err).Error("failed to start transaction")
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	if _, err := tx.Exec(ctx, query,
		sub.ServiceName,
		sub.Price,
		sub.StartDate,
		sub.EndDate,
		sub.UserID,
		sub.ID,
	); err != nil {

		tx.Rollback(ctx)

		d.log.WithError(err).WithFields(logrus.Fields{
			"user_id": sub.UserID,
			"name":    sub.ServiceName,
		}).Error("failed to update subscription")

		return fmt.Errorf("failed to update subscription: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		d.log.WithError(err).Error("failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (d Datasource) DeleteSubscription(ctx context.Context, subID int) error {
	d.log.WithField("subscription_id", subID).Debug("deleting subscription")

	tx, err := d.pool.Begin(ctx)
	if err != nil {
		d.log.WithError(err).Error("failed to start transaction")
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	if _, err := tx.Exec(ctx, "DELETE FROM subscriptions WHERE id = $1", subID); err != nil {
		tx.Rollback(ctx)

		d.log.WithError(err).WithField("subscription_id", subID).
			Error("failed to delete subscription")

		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		d.log.WithError(err).Error("failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (d Datasource) GetSubscriptionsSum(
	ctx context.Context,
	userID, subName string,
	fromDate, toDate time.Time,
) (int, error) {

	query := `SELECT COALESCE(SUM(price), 0) as price
	          FROM subscriptions
	          WHERE user_id = $1 AND name = $2 AND start_date BETWEEN $3 AND $4`

	d.log.WithFields(logrus.Fields{
		"user_id":   userID,
		"name":      subName,
		"from_date": fromDate,
		"to_date":   toDate,
	}).Debug("calculating subscription sum")

	row := d.pool.QueryRow(ctx, query, userID, subName, fromDate, toDate)

	var sum int
	if err := row.Scan(&sum); err != nil {
		d.log.WithError(err).Error("failed to get subscriptions sum")
		return 0, fmt.Errorf("failed to get subscriptions sum: %w", err)
	}

	d.log.WithFields(logrus.Fields{
		"user_id": userID,
		"name":    subName,
		"sum":     sum,
	}).Info("subscription sum calculated")

	return sum, nil
}

func (d Datasource) Stop(_ context.Context) error {
	d.log.Info("closing database connection pool")
	d.pool.Close()
	return nil
}
