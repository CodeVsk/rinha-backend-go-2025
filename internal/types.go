package internal

import "errors"

var ErrFailureToProcessPayment = errors.New("failure to process payment")
var ErrConflictProcess = errors.New("entity already exist")

type PaymentRequest struct {
	CorrelationID string  `json:"correlationId"`
	Amount        float64 `json:"amount"`
}

type PaymentProcessorRequest struct {
	Amount        float64 `json:"amount"`
	CorrelationID string  `json:"correlationId"`
	ProcessedBy   string  `json:"processedBy"`
	RequestedAt   string  `json:"requestedAt"`
}

type PaymentSummaryItem struct {
	TotalAmount   float64 `json:"totalAmount"`
	TotalRequests int     `json:"totalRequests"`
}

type PaymentSummaryResponse struct {
	Default  PaymentSummaryItem `json:"default"`
	Fallback PaymentSummaryItem `json:"fallback"`
}

type HealthCheckResponse struct {
	MinResponseTime int  `json:"minResponseTime"`
	Failing         bool `json:"failing"`
}

type HealthCheckStatus struct {
	Default  HealthCheckResponse `json:"default"`
	Fallback HealthCheckResponse `json:"fallback"`
}
