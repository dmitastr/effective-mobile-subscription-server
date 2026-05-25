package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"effective-mobile-subscription-server/internal/domain/models"
	"effective-mobile-subscription-server/internal/repo"
	"github.com/sirupsen/logrus"
)

var ErrNotFoundById = errors.New("failed to find subscriptions by id")

type IService interface {
	AddSubscription(ctx context.Context, subscription *models.Subscription) error
	GetSubscriptionByUser(ctx context.Context, userID string) ([]models.Subscription, error)
	GetSubscriptionByName(ctx context.Context, subName string) ([]models.Subscription, error)
	UpdateSubscription(ctx context.Context, sub *models.Subscription) error
	DeleteSubscription(ctx context.Context, subID int) error
	GetSubscriptionsSum(ctx context.Context, userID, subName string, fromDate, toDate time.Time) (int, error)
}

type Service struct {
	db  repo.IDatasource
	log *logrus.Logger
}

func NewService(db repo.IDatasource, logger *logrus.Logger) IService {
	return &Service{db: db, log: logger}
}

func (s Service) GetSubscriptionsSum(
	ctx context.Context,
	userID, subName string,
	fromDate, toDate time.Time,
) (int, error) {

	s.log.WithFields(logrus.Fields{
		"user_id":   userID,
		"sub_name":  subName,
		"from_date": fromDate,
		"to_date":   toDate,
	}).Debug("GetSubscriptionsSum called")

	sum, err := s.db.GetSubscriptionsSum(ctx, userID, subName, fromDate, toDate)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"user_id":  userID,
			"sub_name": subName,
		}).Error("failed to get subscriptions sum")

		return 0, fmt.Errorf("GetSubscriptionsSum error: %w", err)
	}

	return sum, nil
}

func (s Service) AddSubscription(ctx context.Context, subscription *models.Subscription) error {
	s.log.WithFields(logrus.Fields{
		"user_id": subscription.UserID,
		"name":    subscription.ServiceName,
	}).Debug("AddSubscription called")

	if err := s.db.AddSubscription(ctx, subscription); err != nil {
		s.log.WithError(err).WithField("user_id", subscription.UserID).
			Error("failed to add subscription")

		return fmt.Errorf("AddSubscription error: %w", err)
	}

	return nil
}

func (s Service) GetSubscriptionByUser(ctx context.Context, userID string) ([]models.Subscription, error) {
	s.log.WithField("user_id", userID).
		Debug("GetSubscriptionByUser called")

	subs, err := s.db.GetSubscriptionByUser(ctx, userID)
	if err != nil {
		s.log.WithError(err).WithField("user_id", userID).
			Error("failed to get subscriptions by user")

		return nil, fmt.Errorf("GetSubscriptionByUser error: %w", err)
	}

	return subs, nil
}

func (s Service) GetSubscriptionByName(ctx context.Context, subName string) ([]models.Subscription, error) {
	s.log.WithField("sub_name", subName).
		Debug("GetSubscriptionByName called")

	subs, err := s.db.GetSubscriptionByName(ctx, subName)
	if err != nil {
		s.log.WithError(err).WithField("sub_name", subName).
			Error("failed to get subscriptions by name")

		return nil, fmt.Errorf("GetSubscriptionByName error: %w", err)
	}

	return subs, nil
}

func (s Service) UpdateSubscription(ctx context.Context, sub *models.Subscription) error {
	s.log.WithFields(logrus.Fields{
		"id":      sub.ID,
		"user_id": sub.UserID,
		"name":    sub.ServiceName,
	}).Debug("UpdateSubscription called")

	subs, err := s.db.GetSubscriptionByID(ctx, sub.ID)
	if err != nil {
		s.log.WithError(err).WithField("id", sub.ID).Error("failed to get subscriptions by id")
		return fmt.Errorf("GetSubscriptionByID error: %w", err)
	}
	if len(subs) != 1 {
		s.log.WithField("id", sub.ID).Error("failed to find subscriptions by id")

		return ErrNotFoundById
	}

	return s.db.UpdateSubscription(ctx, sub)
}

func (s Service) DeleteSubscription(ctx context.Context, subID int) error {
	s.log.WithField("subscription_id", subID).
		Debug("DeleteSubscription called")

	subs, err := s.db.GetSubscriptionByID(ctx, subID)
	if err != nil {
		s.log.WithError(err).WithField("id", subID).Error("failed to get subscriptions by id")
		return fmt.Errorf("GetSubscriptionByID error: %w", err)
	}
	if len(subs) != 1 {
		s.log.WithField("id", subID).Error("failed to find subscriptions by id")

		return ErrNotFoundById
	}

	if err := s.db.DeleteSubscription(ctx, subID); err != nil {
		s.log.WithError(err).WithField("subscription_id", subID).
			Error("failed to delete subscription")

		return fmt.Errorf("DeleteSubscription error: %w", err)
	}

	return nil
}
