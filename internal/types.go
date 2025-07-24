package internal

type PaymentRequest struct {
	CorrelationID string  `json:"correlationId"`
	Amount        float64 `json:"amount"`
}

type PaymentProcessorRequest struct {
	CorrelationID string  `json:"correlationId"`
	Amount        float64 `json:"amount"`
	ProcessedBy   string  `json:"processedBy"`
	RequestedAt   string  `json:"requestedAt"` // ISO 8601 format
}

type PaymentSummaryItem struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type PaymentSummaryResponse struct {
	Default  PaymentSummaryItem `json:"default"`
	Fallback PaymentSummaryItem `json:"fallback"`
}

type HealthCheckResponse struct {
	Failing         bool `json:"failing"`
	MinResponseTime int  `json:"minResponseTime"`
}

type HealthCheckStatus struct {
	Default  HealthCheckResponse `json:"default"`
	Fallback HealthCheckResponse `json:"fallback"`
}
