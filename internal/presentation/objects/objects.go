package objects

type SubscriptionPayload struct {
	ID          int    `json:"id" db:"id"`
	ServiceName string `json:"service_name" db:"name" example:"yandex"`
	Price       int    `json:"price" db:"price" example:"100"`
	UserID      string `json:"user_id" db:"user_id" example:"019e5b65-04a0-7833-bd67-856ea7d20900"`
	StartDate   string `json:"start_date" db:"start_date" example:"10-2020"`
	EndDate     string `json:"end_date" db:"end_date" example:"11-2021"`
}

type AddSubscriptionRequest struct {
	Subscription SubscriptionPayload `json:"subscription"`
}

type AddSubscriptionResponse struct {
}
type UpdateSubscriptionResponse struct {
}

type DeleteSubscriptionResponse struct {
}

type GetSubscriptionResponse struct {
	Subscriptions []SubscriptionPayload `json:"subscriptions"`
}

type SubscriptionSumResponse struct {
	UserID           string `json:"user_id" example:"019e5b65-04a0-7833-bd67-856ea7d20900"`
	SubscriptionName string `json:"subscription" example:"yandex"`
	Sum              int    `json:"sum" example:"10"`
	FromDate         string `json:"from_date" example:"2020-10-01"`
	ToDate           string `json:"to_date" example:"2021-11-01"`
}
type ErrorResponse struct {
}
