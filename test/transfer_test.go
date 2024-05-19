package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"internal-transfers-system/internal/model"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateTransferEdgeCases(t *testing.T) {
	svr := setupTestServer()
	defer svr.DB.Migrator().DropTable(&model.Account{}, &model.Transfer{})

	tests := []struct {
		name               string
		srcAccountID       uint64
		dstAccountID       uint64
		initialSrcBalance  string
		initialDstBalance  string
		payload            string
		expectedSrcBalance string
		expectedDstBalance string
		statusCode         int
	}{
		{
			name:               "Very large transfer amount",
			srcAccountID:       1,
			dstAccountID:       2,
			initialSrcBalance:  "1000000000000000000.00",
			initialDstBalance:  "0.00",
			payload:            `{"source_account_id": 1, "destination_account_id": 2, "amount": "999999999999999999.99"}`,
			expectedSrcBalance: "0.01",
			expectedDstBalance: "999999999999999999.99",
			statusCode:         fiber.StatusCreated,
		},
		{
			name:               "Very small transfer amount",
			srcAccountID:       3,
			dstAccountID:       4,
			initialSrcBalance:  "1000000000000000000.00",
			initialDstBalance:  "0.00",
			payload:            `{"source_account_id": 3, "destination_account_id": 4, "amount": "0.000000000000000001"}`,
			expectedSrcBalance: "999999999999999999.999999999999999999",
			expectedDstBalance: "0.000000000000000001",
			statusCode:         fiber.StatusCreated,
		},
		{
			name:               "Zero transfer amount",
			srcAccountID:       5,
			dstAccountID:       6,
			initialSrcBalance:  "1000000000000000000.00",
			initialDstBalance:  "0.00",
			payload:            `{"source_account_id": 5, "destination_account_id": 6, "amount": "0.00"}`,
			expectedSrcBalance: "1000000000000000000.00",
			expectedDstBalance: "0.00",
			statusCode:         fiber.StatusBadRequest,
		},
		{
			name:               "High precision transfer amount",
			srcAccountID:       7,
			dstAccountID:       8,
			initialSrcBalance:  "1000000000000000000.00",
			initialDstBalance:  "0.00",
			payload:            `{"source_account_id": 7, "destination_account_id": 8, "amount": "0.123456789123456789"}`,
			expectedSrcBalance: "999999999999999999.876543210876543211",
			expectedDstBalance: "0.123456789123456789",
			statusCode:         fiber.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create initial accounts for this test
			svr.DB.Create(&model.Account{ID: tt.srcAccountID, Balance: decimal.RequireFromString(tt.initialSrcBalance)})
			svr.DB.Create(&model.Account{ID: tt.dstAccountID, Balance: decimal.RequireFromString(tt.initialDstBalance)})

			req := httptest.NewRequest("POST", "/transactions", strings.NewReader(tt.payload))
			req.Header.Set("Content-Type", "application/json")

			resp, err := svr.FiberApp.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.statusCode, resp.StatusCode)

			if tt.statusCode == fiber.StatusCreated {
				var srcAccount, dstAccount model.Account
				svr.DB.First(&srcAccount, tt.srcAccountID)
				svr.DB.First(&dstAccount, tt.dstAccountID)

				expectedSrcBalance, _ := decimal.NewFromString(tt.expectedSrcBalance)
				expectedDstBalance, _ := decimal.NewFromString(tt.expectedDstBalance)

				assert.True(t, expectedSrcBalance.Equal(srcAccount.Balance), "expected %v but got %v", expectedSrcBalance, srcAccount.Balance)
				assert.True(t, expectedDstBalance.Equal(dstAccount.Balance), "expected %v but got %v", expectedDstBalance, dstAccount.Balance)
			}
		})
	}
}
