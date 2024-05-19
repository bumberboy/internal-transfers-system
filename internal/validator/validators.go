package validator

import (
	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
	"internal-transfers-system/internal/apimodel"
	"internal-transfers-system/internal/svrerror"
)

func ValidateCreateAccount(account *apimodel.CreateAccountRequest) (decimal.Decimal, error) {
	//Ensure that account ID is greater than 0
	if account.AccountID < 1 {
		return decimal.Zero, svrerror.New("account id must be greater than 0", fiber.StatusBadRequest)
	}

	// Validate initial balance
	initialBalance, err := decimal.NewFromString(account.InitialBalance)
	if err != nil {
		return decimal.Zero, svrerror.New("invalid initial balance", fiber.StatusBadRequest)
	}

	if initialBalance.LessThan(decimal.Zero) {
		return decimal.Zero, svrerror.New("initial balance must be non-negative", fiber.StatusBadRequest)
	}

	return initialBalance, nil
}

func ValidateTransfer(transfer *apimodel.TransferRequest) (decimal.Decimal, error) {
	// Validate amount
	amount, err := decimal.NewFromString(transfer.Amount)
	if err != nil {
		return decimal.Zero, svrerror.New("invalid amount format", fiber.StatusBadRequest)
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, svrerror.New("amount must be greater than zero", fiber.StatusBadRequest)
	}

	// Check for self-transfer
	if transfer.SourceAccountID == transfer.DestinationAccountID {
		return decimal.Zero, svrerror.New("source and destination accounts must be different", fiber.StatusBadRequest)
	}

	return amount, nil
}
