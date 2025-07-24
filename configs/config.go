package configs

import (
	"strconv"

	"github.com/codevsk/rinha-backend-go-2025/pkg/env"
)

type Config struct {
	HttpServerPort                 string
	DefaultProcessorUrl            string
	FallbackProcessorUrl           string
	RedisURL                       string
	WorkersCount                   int
	RetryDefault                   int
	HttpDefaultTimeout             int
	HttpFallbackTimeout            int
	ConsecutiveFailuresDefault     int
	ConsecutiveFailuresFallback    int
	CircuitBreakerIntervalDefault  int
	CircuitBreakerIntervalFallback int
	CircuitBreakerTimeoutDefault   int
	CircuitBreakerTimeoutFallback  int

	PaymentQueueName string
	PaymentTableHash string
}

func NewConfig() *Config {
	env.LoadConfig(".env")

	workersCount, err := strconv.Atoi(env.GetEnv("WORKERS_COUNT", "5"))
	if err != nil {
		panic(err)
	}

	retryDefault, err := strconv.Atoi(env.GetEnv("RETRY_DEFAULT", "5"))
	if err != nil {
		panic(err)
	}

	httpDefaultTimeout, err := strconv.Atoi(env.GetEnv("HTTP_DEFAULT_TIMEOUT", "5"))
	if err != nil {
		panic(err)
	}

	httpFallbackTimeout, err := strconv.Atoi(env.GetEnv("HTTP_FALLBACK_TIMEOUT", "5"))
	if err != nil {
		panic(err)
	}

	consecutiveFailuresDefault, err := strconv.Atoi(env.GetEnv("CONSECUTIVE_FAILURES_DEFAULT", "5"))
	if err != nil {
		panic(err)
	}

	consecutiveFailuresFallback, err := strconv.Atoi(env.GetEnv("CONSECUTIVE_FAILURES_FALLBACK", "5"))
	if err != nil {
		panic(err)
	}

	circuitBreakerIntervalDefault, err := strconv.Atoi(env.GetEnv("CIRCUIT_BREAKER_INTERVAL_DEFAULT", "5"))
	if err != nil {
		panic(err)
	}

	circuitBreakerIntervalFallback, err := strconv.Atoi(env.GetEnv("CIRCUIT_BREAKER_INTERVAL_FALLBACK", "5"))
	if err != nil {
		panic(err)
	}

	circuitBreakerTimeoutDefault, err := strconv.Atoi(env.GetEnv("CIRCUIT_BREAKER_TIMEOUT_DEFAULT", "5"))
	if err != nil {
		panic(err)
	}

	circuitBreakerTimeoutFallback, err := strconv.Atoi(env.GetEnv("CIRCUIT_BREAKER_TIMEOUT_FALLBACK", "5"))
	if err != nil {
		panic(err)
	}

	return &Config{
		HttpServerPort:                 env.GetEnv("HTTP_PORT", "9999"),
		DefaultProcessorUrl:            env.GetEnv("DEFAULT_URL", "http://localhost:8001"),
		FallbackProcessorUrl:           env.GetEnv("FALLBACK_URL", "http://localhost:8002"),
		RedisURL:                       env.GetEnv("REDIS_URL", "localhost:6379"),
		PaymentQueueName:               env.GetEnv("PAYMENT_QUEUE_NAME", "payments_pending"),
		PaymentTableHash:               env.GetEnv("PAYMENT_TABLE_HASH", "payments"),
		WorkersCount:                   workersCount,
		RetryDefault:                   retryDefault,
		HttpDefaultTimeout:             httpDefaultTimeout,
		HttpFallbackTimeout:            httpFallbackTimeout,
		ConsecutiveFailuresDefault:     consecutiveFailuresDefault,
		ConsecutiveFailuresFallback:    consecutiveFailuresFallback,
		CircuitBreakerIntervalDefault:  circuitBreakerIntervalDefault,
		CircuitBreakerIntervalFallback: circuitBreakerIntervalFallback,
		CircuitBreakerTimeoutDefault:   circuitBreakerTimeoutDefault,
		CircuitBreakerTimeoutFallback:  circuitBreakerTimeoutFallback,
	}
}
