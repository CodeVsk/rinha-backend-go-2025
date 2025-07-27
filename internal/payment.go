package internal

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/araddon/dateparse"
	"github.com/bytedance/sonic"
	"github.com/codevsk/rinha-backend-go-2025/configs"
	"github.com/go-redis/redis/v8"
)

type PaymentService struct {
	paymentQueue chan PaymentRequest
	restClient   *RestClient
	cfg          *configs.Config
	rdb          *redis.Client
}

func NewPaymentService(restClient *RestClient, cfg *configs.Config, rdb *redis.Client) *PaymentService {
	return &PaymentService{
		paymentQueue: make(chan PaymentRequest, cfg.PaymentQueueChanSize),
		restClient:   restClient,
		cfg:          cfg,
		rdb:          rdb,
	}
}

func (p *PaymentService) EnqueuePayment(payment PaymentRequest) {
	p.paymentQueue <- payment
}

func (p *PaymentService) StartWorkers() {
	for i := 0; i < p.cfg.WorkersCount; i++ {
		go p.worker()
	}
}

func (p *PaymentService) worker() {
	for payment := range p.paymentQueue {
		if err := p.sendPayment(context.Background(), payment); err != nil {
			if !errors.Is(err, ErrConflictProcess) {
				p.paymentQueue <- payment
			}
		}

	}
}

func (p *PaymentService) sendPayment(ctx context.Context, payment PaymentRequest) error {
	paymentRequest := PaymentProcessorRequest{
		CorrelationID: payment.CorrelationID,
		Amount:        payment.Amount,
	}

	processors := []func(PaymentProcessorRequest) ([]byte, error){
		p.restClient.SendPaymentDefault,
		p.restClient.SendPaymentFallback,
	}

	for _, processor := range processors {
		bodyBytes, err := processor(paymentRequest)
		if err != nil {
			continue
		}

		if bodyBytes == nil {
			return ErrConflictProcess
		}

		if err := p.rdb.HSet(ctx, p.cfg.PaymentTableHash, payment.CorrelationID, bodyBytes).Err(); err != nil {
			return fmt.Errorf("failed to save payment in redis: %w", err)
		}

		return nil
	}

	return ErrFailureToProcessPayment

}

func (p *PaymentService) GetPaymentsSummary(ctx context.Context, fromStr string, toStr string) (PaymentSummaryResponse, error) {
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
		if err := sonic.Unmarshal([]byte(payment), &item); err != nil {
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

	output.Default.TotalAmount = math.Round(output.Default.TotalAmount*10) / 10
	output.Fallback.TotalAmount = math.Round(output.Fallback.TotalAmount*10) / 10

	return output, nil
}

func (p *PaymentService) parseTimeRange(fromStr, toStr string) (time.Time, time.Time, bool, error) {
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
