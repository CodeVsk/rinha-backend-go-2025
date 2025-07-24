package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/araddon/dateparse"
	"github.com/codevsk/rinha-backend-go-2025/configs"
	"github.com/go-redis/redis/v8"
)

var ErrFailureToProcessPayment = errors.New("failure to process payment")
var ErrUnavailableProcessor = errors.New("unavailable processors")
var ErrConflictProcess = errors.New("entity already exist")
var ErrFallbackProcess = errors.New("fallback process error")

type Payment struct {
	paymentQueue chan PaymentRequest
	semaphore    chan struct{}
	client       http.Client
	restClient   *RestClient
	cfg          *configs.Config
	rdb          *redis.Client
}

func NewPayment(paymentQueue chan PaymentRequest, client http.Client, restClient *RestClient, cfg *configs.Config, rdb *redis.Client) *Payment {
	return &Payment{
		paymentQueue: paymentQueue,
		semaphore:    make(chan struct{}, cfg.WorkersCount),
		client:       client,
		restClient:   restClient,
		cfg:          cfg,
		rdb:          rdb,
	}
}

func (p *Payment) EnqueuePayment(payment PaymentRequest) {
	p.paymentQueue <- payment
}

func (p *Payment) Worker() {
	for payment := range p.paymentQueue {
		if err := p.ProcessPayment(context.Background(), payment); err != nil {
			p.paymentQueue <- payment
		}
	}
}

func (p *Payment) ProcessPayment(ctx context.Context, payment PaymentRequest) error {
	paymentBytes, err := p.httpPostWithRetry(payment)
	if err != nil {
		if errors.Is(err, ErrConflictProcess) {
			return nil
		}
		return err
	}

	err = p.rdb.HSet(ctx, p.cfg.PaymentTableHash, payment.CorrelationID, paymentBytes).Err()
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("failed to save payment in redis: %w", err)
	}

	return nil
}

func (p *Payment) httpPostWithRetry(payment PaymentRequest) ([]byte, error) {
	var bodyBytes []byte
	var err error

	paymentRequest := PaymentProcessorRequest{
		CorrelationID: payment.CorrelationID,
		Amount:        payment.Amount,
	}

	bodyBytes, err = p.restClient.SendPaymentDefault(paymentRequest)
	if err == nil && bodyBytes == nil {
		return nil, ErrConflictProcess
	}
	if bodyBytes != nil {
		return bodyBytes, nil
	}

	bodyBytes, err = p.restClient.SendPaymentFallback(paymentRequest)
	if err == nil && bodyBytes == nil {
		return nil, ErrConflictProcess
	}
	if bodyBytes != nil {
		return bodyBytes, nil
	}

	return nil, ErrFailureToProcessPayment
}

/*
func (p *Payment) httpPostWithRetry(payment PaymentRequest) ([]byte, error) {
	var paymentRequest PaymentProcessorRequest

	endpoint := p.cfg.DefaultProcessorUrl
	paymentRequest.ProcessedBy = "default"
	paymentRequest.CorrelationID = payment.CorrelationID
	paymentRequest.Amount = payment.Amount

	for attempt := 1; attempt <= p.cfg.RetryDefault; attempt++ {
		paymentRequest.RequestedAt = time.Now().UTC().Format(time.RFC3339Nano)

		bodyBytes, err := json.Marshal(paymentRequest)
		if err != nil {
			return nil, err
		}

		res, err := p.client.Post(endpoint+"/payments", "application/json", bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		if res.StatusCode == 409 || res.StatusCode == 422 {
			fmt.Print(ErrConflictProcess.Error())
			return nil, ErrConflictProcess
		}

		//if res.StatusCode == 500 {
		//	time.Sleep(7 * time.Second)
		//	continue
		//}

		if res.StatusCode >= 200 && res.StatusCode <= 300 {
			return bodyBytes, nil
		}
	}

	paymentRequest.ProcessedBy = "fallback"
	paymentRequest.RequestedAt = time.Now().UTC().Format(time.RFC3339Nano)

	bodyBytes, err := json.Marshal(paymentRequest)
	if err != nil {
		return nil, err
	}

	client2 := http.Client{
		Timeout: time.Duration(p.cfg.HttpFallbackTimeout) * time.Second,
	}
	res, err := client2.Post(p.cfg.FallbackProcessorUrl+"/payments", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	if res.StatusCode == 409 || res.StatusCode == 422 {
		return nil, ErrConflictProcess
	}

	if res.StatusCode >= 200 && res.StatusCode <= 300 {
		return bodyBytes, nil
	}
	defer res.Body.Close()

	return nil, ErrFailureToProcessPayment
}*/

func (p *Payment) GetPaymentsSummary(ctx context.Context, fromStr string, toStr string) (PaymentSummaryResponse, error) {
	paymentList, err := p.rdb.HGetAll(ctx, p.cfg.PaymentTableHash).Result()
	if err != nil {
		return PaymentSummaryResponse{}, err
	}

	output := PaymentSummaryResponse{
		Default: PaymentSummaryItem{
			TotalAmount:   0,
			TotalRequests: 0,
		},
		Fallback: PaymentSummaryItem{
			TotalAmount:   0,
			TotalRequests: 0,
		},
	}

	from, to, useFilter, err := p.parseTimeRange(fromStr, toStr)
	if err != nil {
		return PaymentSummaryResponse{}, err
	}

	for _, payment := range paymentList {
		var item PaymentProcessorRequest
		if err := json.Unmarshal([]byte(payment), &item); err != nil {
			continue
		}

		requestedAt, err := time.Parse(time.RFC3339Nano, item.RequestedAt)

		if useFilter {
			if err != nil || requestedAt.Before(from) || requestedAt.After(to) {
				continue
			}
		}

		switch item.ProcessedBy {
		case "default":
			output.Default.TotalAmount += item.Amount
			output.Default.TotalRequests++
		case "fallback":
			output.Fallback.TotalAmount += item.Amount
			output.Fallback.TotalRequests++
		}
	}

	output.Default.TotalAmount = math.Round(float64(output.Default.TotalAmount)*10) / 10
	output.Fallback.TotalAmount = math.Round(float64(output.Fallback.TotalAmount)*10) / 10

	return output, nil
}

func (p *Payment) parseTimeRange(fromStr, toStr string) (time.Time, time.Time, bool, error) {
	if fromStr == "" && toStr == "" {
		return time.Time{}, time.Time{}, false, nil
	}
	if fromStr == "" || toStr == "" {
		return time.Time{}, time.Time{}, false, fmt.Errorf("both 'from' and 'to' parameters must be provided")
	}
	from, err := dateparse.ParseAny(fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, false, fmt.Errorf("invalid date format: %v", err)
	}
	to, err := dateparse.ParseAny(toStr)
	if err != nil {
		return time.Time{}, time.Time{}, false, fmt.Errorf("invalid date format: %v", err)
	}

	return from, to, true, nil
}
