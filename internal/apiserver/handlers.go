package apiserver

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"internal-transfers-system/internal/apimodel"
	"internal-transfers-system/internal/model"
	"internal-transfers-system/internal/service"
	"internal-transfers-system/internal/svrerror"
	"internal-transfers-system/internal/validator"
	"strings"
)

func (s *Server) CreateAccount(c *fiber.Ctx) error {
	var account apimodel.CreateAccountRequest

	if err := c.BodyParser(&account); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	initialBalance, err := validator.ValidateCreateAccount(&account)
	if err != nil {
		var customErr *svrerror.Error
		if errors.As(err, &customErr) {
			return c.Status(customErr.StatusCode).JSON(fiber.Map{"error": customErr.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	newAccount := model.Account{
		ID:      account.AccountID,
		Balance: initialBalance,
	}

	if err := s.DB.WithContext(c.Context()).Create(&newAccount).Error; err != nil {

		if strings.Contains(err.Error(), "unique") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "account ID already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{})
}

func (s *Server) GetAccount(c *fiber.Ctx) error {
	accountID := c.Params("account_id")

	var account model.Account
	if err := s.DB.WithContext(c.Context()).First(&account, "id = ?", accountID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "account not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	response := struct {
		AccountID uint64 `json:"account_id"`
		Balance   string `json:"balance"`
	}{
		AccountID: account.ID,
		Balance:   account.Balance.String(),
	}

	return c.JSON(response)
}

func (s *Server) CreateTransfer(c *fiber.Ctx) error {
	var transfer apimodel.TransferRequest

	if err := c.BodyParser(&transfer); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	amount, err := validator.ValidateTransfer(&transfer)
	if err != nil {
		var customErr *svrerror.Error
		if errors.As(err, &customErr) {
			return c.Status(customErr.StatusCode).JSON(fiber.Map{"error": customErr.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	err = service.ProcessTransfer(c.Context(), s.DB, transfer, amount)
	if err != nil {
		var customErr *svrerror.Error
		if errors.As(err, &customErr) {
			return c.Status(customErr.StatusCode).JSON(fiber.Map{"error": customErr.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{})
}
