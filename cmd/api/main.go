package main

import (
	"fmt"
	"net/http"
	"time"

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

	pq := make(chan internal.PaymentRequest, 20000)

	c := http.Client{
		Timeout: time.Duration(cfg.HttpDefaultTimeout) * time.Second,
	}

	rc := internal.NewRestClient(*cfg)

	p := internal.NewPayment(pq, c, rc, cfg, db)

	for i := 0; i < cfg.WorkersCount; i++ {
		go p.Worker()
	}

	newFiber(cfg, p)
}

func newFiber(cfg *configs.Config, payment *internal.Payment) {
	f := fiber.New(fiber.Config{
		ServerHeader: "Fiber",
		AppName:      "CodeVSK",
	})

	f.Get("/payments-summary", func(c *fiber.Ctx) error {
		fromStr := c.Query("from")
		toStr := c.Query("to")

		summary, err := payment.GetPaymentsSummary(c.Context(), fromStr, toStr)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(summary)
	})

	f.Post("/payments", func(c *fiber.Ctx) error {
		var input internal.PaymentRequest

		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
		}

		payment.EnqueuePayment(input)

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Payment registered successfully"})
	})

	if err := f.Listen(fmt.Sprintf(":%s", cfg.HttpServerPort)); err != nil {
		panic(err)
	}
}
