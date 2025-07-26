package handler

import (
	"github.com/codevsk/rinha-backend-go-2025/internal"
	"github.com/gofiber/fiber/v2"
)

type PaymentHandler struct {
	PaymentService *internal.PaymentService
}

func NewPaymentHandler(paymentService *internal.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		PaymentService: paymentService,
	}
}

func (h *PaymentHandler) RequestPayment(c *fiber.Ctx) error {
	var input internal.PaymentRequest

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	h.PaymentService.EnqueuePayment(input)

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Payment registered successfully"})
}

func (h *PaymentHandler) GetPaymentsSummary(c *fiber.Ctx) error {
	fromStr := c.Query("from")
	toStr := c.Query("to")

	summary, err := h.PaymentService.GetPaymentsSummary(c.Context(), fromStr, toStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(summary)
}
