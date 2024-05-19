package validator

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"internal-transfers-system/internal/apimodel"
	"internal-transfers-system/internal/svrerror"
	"testing"
)

func TestValidateTransfer(t *testing.T) {
	tests := []struct {
		name           string
		transfer       apimodel.TransferRequest
		expectedError  error
		expectedAmount decimal.Decimal
	}{
		{
			name: "valid transfer",
			transfer: apimodel.TransferRequest{
				SourceAccountID:      1,
				DestinationAccountID: 2,
				Amount:               "100.50",
			},
			expectedError:  nil,
			expectedAmount: decimal.NewFromFloat(100.50),
		},
		{
			name: "invalid amount format",
			transfer: apimodel.TransferRequest{
				SourceAccountID:      1,
				DestinationAccountID: 2,
				Amount:               "invalid",
			},
			expectedError:  svrerror.New("invalid amount format", fiber.StatusBadRequest),
			expectedAmount: decimal.Zero,
		},
		{
			name: "amount less than or equal to zero",
			transfer: apimodel.TransferRequest{
				SourceAccountID:      1,
				DestinationAccountID: 2,
				Amount:               "0",
			},
			expectedError:  svrerror.New("amount must be greater than zero", fiber.StatusBadRequest),
			expectedAmount: decimal.Zero,
		},
		{
			name: "self-transfer",
			transfer: apimodel.TransferRequest{
				SourceAccountID:      1,
				DestinationAccountID: 1,
				Amount:               "100.50",
			},
			expectedError:  svrerror.New("source and destination accounts must be different", fiber.StatusBadRequest),
			expectedAmount: decimal.Zero,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, err := ValidateTransfer(&tt.transfer)

			if tt.expectedError != nil {
				assert.Error(t, err)
				var customErr *svrerror.Error
				ok := errors.As(err, &customErr)
				var expectedError *svrerror.Error
				_ = errors.As(tt.expectedError, &expectedError)
				assert.True(t, ok)
				assert.Equal(t, expectedError.Message, customErr.Message)
				assert.Equal(t, expectedError.StatusCode, customErr.StatusCode)
			} else {
				assert.NoError(t, err)
			}
			assert.True(t, amount.Equal(tt.expectedAmount))
		})
	}
}

func TestValidateCreateAccount(t *testing.T) {
	tests := []struct {
		name          string
		account       *apimodel.CreateAccountRequest
		expectedValue decimal.Decimal
		expectedError error
	}{
		{
			name: "Valid initial balance",
			account: &apimodel.CreateAccountRequest{
				AccountID:      1,
				InitialBalance: "100",
			},
			expectedValue: decimal.NewFromFloat(100.00),
			expectedError: nil,
		},
		{
			name: "invalid account ID",
			account: &apimodel.CreateAccountRequest{
				AccountID:      0,
				InitialBalance: "100",
			},
			expectedValue: decimal.Zero,
			expectedError: svrerror.New("account id must be greater than 0", fiber.StatusBadRequest),
		},
		{
			name: "Invalid initial balance format",
			account: &apimodel.CreateAccountRequest{
				AccountID:      2,
				InitialBalance: "invalid",
			},
			expectedValue: decimal.Zero,
			expectedError: svrerror.New("invalid initial balance", fiber.StatusBadRequest),
		},
		{
			name: "Negative initial balance",
			account: &apimodel.CreateAccountRequest{
				AccountID:      3,
				InitialBalance: "-100",
			},
			expectedValue: decimal.Zero,
			expectedError: svrerror.New("initial balance must be non-negative", fiber.StatusBadRequest),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := ValidateCreateAccount(tt.account)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			assert.True(t, tt.expectedValue.Equal(value), "expected %v but got %v", tt.expectedValue, value)
		})
	}
}
