package internal

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/codevsk/rinha-backend-go-2025/configs"
	"github.com/sony/gobreaker/v2"
)

type RestClient struct {
	cfg            configs.Config
	defaultClient  http.Client
	fallbackClient http.Client
	cbDefault      *gobreaker.CircuitBreaker[[]byte]
	cbFallback     *gobreaker.CircuitBreaker[[]byte]
}

func NewRestClient(cfg configs.Config) *RestClient {
	return &RestClient{
		cfg:            cfg,
		defaultClient:  http.Client{Timeout: time.Duration(cfg.HttpDefaultTimeout) * time.Second},
		fallbackClient: http.Client{Timeout: time.Duration(cfg.HttpFallbackTimeout) * time.Second},
		cbDefault: gobreaker.NewCircuitBreaker[[]byte](gobreaker.Settings{
			Name:        "DefaultProcessor",
			MaxRequests: uint32(cfg.RetryDefault),
			Interval:    time.Duration(cfg.CircuitBreakerIntervalDefault) * time.Millisecond,
			Timeout:     time.Duration(cfg.CircuitBreakerTimeoutDefault) * time.Second, //time.Duration(cfg.HttpDefaultTimeout) * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > uint32(cfg.ConsecutiveFailuresDefault)
			},
		}),
		cbFallback: gobreaker.NewCircuitBreaker[[]byte](gobreaker.Settings{
			Name:        "FallbackProcessor",
			MaxRequests: 1,
			Interval:    time.Duration(cfg.CircuitBreakerIntervalFallback) * time.Millisecond,
			Timeout:     time.Duration(cfg.CircuitBreakerTimeoutFallback) * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > uint32(cfg.ConsecutiveFailuresFallback)
			},
		}),
	}
}

func (r *RestClient) SendPaymentDefault(paymentRequest PaymentProcessorRequest) (bodyBytes []byte, err error) {
	bodyBytes, err = r.cbDefault.Execute(func() ([]byte, error) {
		paymentRequest.ProcessedBy = "default"
		paymentRequest.RequestedAt = time.Now().UTC().Format(time.RFC3339Nano)

		bodyBytes, err := sonic.Marshal(paymentRequest)
		if err != nil {
			return nil, err
		}

		res, err := r.defaultClient.Post(r.cfg.DefaultProcessorUrl+"/payments", "application/json", bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		_, err = io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		if res.StatusCode == 409 || res.StatusCode == 422 {
			return nil, nil
		}

		if res.StatusCode >= 200 && res.StatusCode <= 300 {
			return bodyBytes, nil
		}

		return nil, ErrFailureToProcessPayment
	})

	return
}

func (r *RestClient) SendPaymentFallback(paymentRequest PaymentProcessorRequest) (bodyBytes []byte, err error) {
	bodyBytes, err = r.cbFallback.Execute(func() ([]byte, error) {
		paymentRequest.ProcessedBy = "fallback"
		paymentRequest.RequestedAt = time.Now().UTC().Format(time.RFC3339Nano)

		bodyBytes, err := sonic.Marshal(paymentRequest)
		if err != nil {
			return nil, err
		}

		res, err := r.fallbackClient.Post(r.cfg.FallbackProcessorUrl+"/payments", "application/json", bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		_, err = io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		if res.StatusCode == 409 || res.StatusCode == 422 {
			return nil, nil
		}

		if res.StatusCode >= 200 && res.StatusCode <= 300 {
			return bodyBytes, nil
		}

		return nil, ErrFailureToProcessPayment
	})

	return
}
