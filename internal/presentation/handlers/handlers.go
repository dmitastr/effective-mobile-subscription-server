package handlers

import (
	"net/http"
	"strconv"
	"time"

	"effective-mobile-subscription-server/internal/domain/models"
	subService "effective-mobile-subscription-server/internal/domain/service"
	"effective-mobile-subscription-server/internal/mapper"
	"effective-mobile-subscription-server/internal/presentation/objects"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type IHandler interface {
	AddSubscription(*gin.Context)
	GetSubscription(*gin.Context)
	UpdateSubscription(*gin.Context)
	DeleteSubscription(*gin.Context)
	GetSubscriptionAggregate(*gin.Context)
}

type Handler struct {
	service subService.IService
	logger  *logrus.Logger
}

func NewHandler(service subService.IService, log *logrus.Logger) IHandler {
	return &Handler{service: service, logger: log}
}

// AddSubscription godoc
// @Summary Add new subscription
// @Description Add new subscription
// @Tags subscription
// @Produce json
// @Param	subscription	body objects.AddSubscriptionRequest	true	"SubscriptionPayload payload"
// @Success 200 {object} objects.AddSubscriptionResponse
// @Failure 400 {object} objects.ErrorResponse
// @Failure 500 {object} objects.ErrorResponse
// @Router /subscription [post]
func (h Handler) AddSubscription(c *gin.Context) {
	var request objects.AddSubscriptionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.WithError(err).Error("failed to parse request body")

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"subscription": request,
	}).Info("add subscription request")

	subscription, err := mapper.ToModel(request.Subscription)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"subscription": request,
		}).Error("failed to convert to model")

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddSubscription(c, subscription); err != nil {

		h.logger.WithError(err).WithFields(logrus.Fields{
			"subscriptionModel": subscription,
		}).Error("failed to add subscription")

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"subscriptionModel": subscription,
	}).Info("subscription added successfully")

	c.JSON(http.StatusOK, objects.AddSubscriptionResponse{})
}

// GetSubscription godoc
// @Summary Get subscriptions
// @Description Get subscriptions by user_id or sub name
// @Tags subscription
// @Produce json
// @Param	user_id	query string	false	"user id in UUID string format"
// @Param	subscription_name	query string	false	"subscription name"
// @Success 200 {object} objects.GetSubscriptionResponse
// @Failure 400 {object} objects.ErrorResponse
// @Failure 500 {object} objects.ErrorResponse
// @Router /subscription [get]
func (h Handler) GetSubscription(c *gin.Context) {
	userID := c.Query("user_id")
	subName := c.Query("subscription_name")

	h.logger.WithFields(logrus.Fields{
		"user_id":           userID,
		"subscription_name": subName,
	}).Info("get subscription request")

	var subs []models.Subscription
	var err error
	if userID == "" && subName == "" {
		h.logger.Warn("get subscription: missing user_id and subscription_name")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing user_id or subscription_name"})
		return
	} else if userID != "" {
		subs, err = h.service.GetSubscriptionByUser(c, userID)
	} else {
		subs, err = h.service.GetSubscriptionByName(c, subName)
	}

	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":           userID,
			"subscription_name": subName,
		}).Error("failed to get subscriptions")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":           userID,
		"subscription_name": subName,
		"count":             len(subs),
	}).Info("subscriptions retrieved")

	c.JSON(http.StatusOK, objects.GetSubscriptionResponse{Subscriptions: mapper.ToDTO(subs...)})
}

// UpdateSubscription godoc
// @Summary update subscription data
// @Description  update subscription data
// @Tags subscription
// @Produce json
// @Param id path string true "subscription id"
// @Param	subscription	body objects.AddSubscriptionRequest	true	"SubscriptionPayload payload"
// @Success 200 {object} objects.UpdateSubscriptionResponse
// @Failure 400 {object} objects.ErrorResponse
// @Failure 500 {object} objects.ErrorResponse
// @Router /subscription/{id} [put]
func (h Handler) UpdateSubscription(c *gin.Context) {
	subID := c.Param("id")

	h.logger.WithFields(logrus.Fields{
		"subscriptionID": subID,
	}).Info("update subscription request")

	if subID == "" {
		h.logger.Warn("update subscription: missing subscription ID")

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing subscription id"})
		return
	}
	var request objects.AddSubscriptionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"subscription": request,
		}).Error("update subscription: failed to parse request body")

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.Atoi(subID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"subscriptionID": subID,
		}).Error("update subscription: invalid subscription ID")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	request.Subscription.ID = id

	subscription, err := mapper.ToModel(request.Subscription)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"subscription": request,
		}).Error("update subscription: failed to convert to model")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateSubscription(c, subscription); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"subscriptionModel": subscription,
		}).Error("update subscription: failed to update subscription")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"subscriptionModel": subscription,
	}).Info("subscription updated successfully")

	c.JSON(http.StatusOK, objects.UpdateSubscriptionResponse{})
}

// DeleteSubscription godoc
// @Summary delete subscription
// @Description  delete subscription by id
// @Tags subscription
// @Produce json
// @Param id path string true "subscription id"
// @Success 200 {object} objects.DeleteSubscriptionResponse
// @Failure 400 {object} objects.ErrorResponse
// @Failure 500 {object} objects.ErrorResponse
// @Router /subscription/{id} [delete]
func (h Handler) DeleteSubscription(c *gin.Context) {
	subID := c.Param("id")

	h.logger.WithField("subscription_id", subID).
		Info("delete subscription request")

	if subID == "" {
		h.logger.Warn("delete subscription: missing subscription id")

		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "missing subscription id"})
		return
	}

	id, err := strconv.Atoi(subID)
	if err != nil {
		h.logger.WithError(err).WithField("subscription_id", subID).
			Warn("delete subscription: invalid subscription id")

		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}

	if err := h.service.DeleteSubscription(c, id); err != nil {
		h.logger.WithError(err).WithField("subscription_id", id).
			Error("failed to delete subscription")

		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	h.logger.WithField("subscription_id", id).
		Info("subscription deleted")

	c.JSON(http.StatusOK, objects.DeleteSubscriptionResponse{})
}

// GetSubscriptionAggregate godoc
// @Summary get aggregate data
// @Description get sum of specified subscriptions for a user in certain period
// @Tags subscription
// @Produce json
// @Param user_id query string true "user id in uuid format"
// @Param subscription query string true "subscription name"
// @Param from_date query string true "Min starting date of subscription (YYYY-MM-DD)"	format(date) example(2025-08-02)
// @Param to_date query string true	  "Min starting date of subscription (YYYY-MM-DD)"	format(date) example(2025-08-02)
// @Success 200 {object} objects.SubscriptionSumResponse
// @Failure 400 {object} objects.ErrorResponse
// @Failure 500 {object} objects.ErrorResponse
// @Router /subscription/aggregate [get]
func (h Handler) GetSubscriptionAggregate(c *gin.Context) {
	userID := c.Query("user_id")
	subName := c.Query("subscription")
	from := c.Query("from_date")
	to := c.Query("to_date")

	h.logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"subscription": subName,
		"from_date":    from,
		"to_date":      to,
	}).Info("get subscription aggregate request")

	if userID == "" || subName == "" || from == "" || to == "" {
		h.logger.Warn("get subscription aggregate: missing required query parameters")

		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "missing user_id or subscription id or from/to date"})
		return
	}

	fromDate, errFrom := time.Parse("2006-01-02", from)
	toDate, errTo := time.Parse("2006-01-02", to)

	var errDates string
	if errFrom != nil {
		errDates += errFrom.Error() + "\n"
	}
	if errTo != nil {
		errDates += errTo.Error() + "\n"
	}

	if errDates != "" {
		h.logger.WithFields(logrus.Fields{
			"user_id":      userID,
			"subscription": subName,
			"from_date":    from,
			"to_date":      to,
		}).Warn("invalid date format")

		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": errDates})
		return
	}

	sum, err := h.service.GetSubscriptionsSum(
		c,
		userID,
		subName,
		fromDate,
		toDate,
	)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":      userID,
			"subscription": subName,
			"from_date":    from,
			"to_date":      to,
		}).Error("failed to get subscription aggregate")

		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"subscription": subName,
		"sum":          sum,
	}).Info("subscription aggregate calculated")

	c.JSON(http.StatusOK, objects.SubscriptionSumResponse{
		UserID:           userID,
		SubscriptionName: subName,
		Sum:              sum,
		FromDate:         from,
		ToDate:           to,
	})
}
