package mapper

import (
	"time"

	"effective-mobile-subscription-server/internal/core/common"
	"effective-mobile-subscription-server/internal/domain/models"
	"effective-mobile-subscription-server/internal/presentation/objects"

	"github.com/google/uuid"
)

func ToModel(dto objects.SubscriptionPayload) (*models.Subscription, error) {
	if _, err := uuid.Parse(dto.UserID); err != nil {
		return nil, err
	}
	startDate, err := time.Parse(common.DateLayout, dto.StartDate)
	if err != nil {
		return nil, err
	}

	endDate, err := time.Parse(common.DateLayout, dto.EndDate)
	if err != nil {
		return nil, err
	}

	return &models.Subscription{
		ID:          dto.ID,
		ServiceName: dto.ServiceName,
		Price:       dto.Price,
		UserID:      dto.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
	}, nil
}

func ToDTO(subscriptionModels ...models.Subscription) []objects.SubscriptionPayload {
	dtos := make([]objects.SubscriptionPayload, len(subscriptionModels))
	for i, m := range subscriptionModels {
		dtos[i] = objects.SubscriptionPayload{
			ID:          m.ID,
			ServiceName: m.ServiceName,
			Price:       m.Price,
			UserID:      m.UserID,
			StartDate:   m.StartDate.Format(common.DateLayout),
			EndDate:     m.EndDate.Format(common.DateLayout),
		}
	}

	return dtos
}
