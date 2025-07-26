package main

import (
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/codevsk/rinha-backend-go-2025/cmd/api/handler"
	"github.com/codevsk/rinha-backend-go-2025/configs"
	"github.com/codevsk/rinha-backend-go-2025/internal"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg := configs.NewConfig()

	db := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})

	c := internal.NewRestClient(*cfg)

	p := internal.NewPaymentService(c, cfg, db)

	p.StartWorkers()

	h := handler.NewPaymentHandler(p)

	f := fiber.New(fiber.Config{
		ServerHeader: "Fiber",
		AppName:      "CodeVSK",
		JSONEncoder:  sonic.Marshal,
		JSONDecoder:  sonic.Unmarshal,
	})

	f.Get("/payments-summary", h.GetPaymentsSummary)

	f.Post("/payments", h.RequestPayment)

	if err := f.Listen(fmt.Sprintf(":%s", cfg.HttpServerPort)); err != nil {
		panic(err)
	}
}
