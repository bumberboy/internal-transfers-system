package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"internal-transfers-system/internal/model"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestConcurrentTransfers(t *testing.T) {
	svr := setupTestServer()
	defer svr.DB.Migrator().DropTable(&model.Account{}, &model.Transfer{})

	// Create initial accounts
	svr.DB.Create(&model.Account{ID: 1, Balance: decimal.NewFromFloat(1000.00)})
	svr.DB.Create(&model.Account{ID: 2, Balance: decimal.NewFromFloat(1000.00)})

	payload := `{"source_account_id": 1, "destination_account_id": 2, "amount": "10.00"}`

	transferFunc := func() {
		req := httptest.NewRequest("POST", "/transactions", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		resp, err := svr.FiberApp.Test(req, 5000)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	}

	const numTransfers = 10
	var wg sync.WaitGroup
	for i := 0; i < numTransfers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			transferFunc()
			t.Log("done transferring", "transfer ID", i)
		}()
	}
	wg.Wait()

	// Verify the final balances
	var sourceAccount, destinationAccount model.Account
	svr.DB.First(&sourceAccount, 1)
	svr.DB.First(&destinationAccount, 2)

	expectedSourceBalance := decimal.NewFromFloat(1000.00).Sub(decimal.NewFromFloat(10.00).Mul(decimal.NewFromFloat(numTransfers)))
	expectedDestinationBalance := decimal.NewFromFloat(1000.00).Add(decimal.NewFromFloat(10.00).Mul(decimal.NewFromFloat(numTransfers)))

	assert.True(t, expectedSourceBalance.Equal(sourceAccount.Balance), "expected %v but got %v", expectedSourceBalance, sourceAccount.Balance)
	assert.True(t, expectedDestinationBalance.Equal(destinationAccount.Balance), "expected %v but got %v", expectedDestinationBalance, destinationAccount.Balance)
}

func TestConcurrentTransfersDifferentAccounts(t *testing.T) {
	svr := setupTestServer()
	defer svr.DB.Migrator().DropTable(&model.Account{}, &model.Transfer{})

	// Create initial accounts
	svr.DB.Create(&model.Account{ID: 1, Balance: decimal.NewFromFloat(1000.00)})
	svr.DB.Create(&model.Account{ID: 2, Balance: decimal.NewFromFloat(1000.00)})
	svr.DB.Create(&model.Account{ID: 3, Balance: decimal.NewFromFloat(1000.00)})
	svr.DB.Create(&model.Account{ID: 4, Balance: decimal.NewFromFloat(1000.00)})

	payload1 := `{"source_account_id": 1, "destination_account_id": 2, "amount": "10.00"}`
	payload2 := `{"source_account_id": 3, "destination_account_id": 4, "amount": "20.00"}`

	transferFunc := func(payload string) {
		req := httptest.NewRequest("POST", "/transactions", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		resp, err := svr.FiberApp.Test(req, 5000)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	}

	const numTransfers = 10
	var wg sync.WaitGroup
	for i := 0; i < numTransfers; i++ {
		wg.Add(1)
		go func(payload string) {
			defer wg.Done()
			transferFunc(payload)
			t.Log("done transferring", "transfer ID", i)
		}(payload1)
		wg.Add(1)
		go func(payload string) {
			defer wg.Done()
			transferFunc(payload)
			t.Log("done transferring", "transfer ID", i)
		}(payload2)
	}
	wg.Wait()

	// Verify the final balances
	var account1, account2, account3, account4 model.Account
	svr.DB.First(&account1, 1)
	svr.DB.First(&account2, 2)
	svr.DB.First(&account3, 3)
	svr.DB.First(&account4, 4)

	expectedBalance1 := decimal.NewFromFloat(1000.00).Sub(decimal.NewFromFloat(10.00).Mul(decimal.NewFromFloat(numTransfers)))
	expectedBalance2 := decimal.NewFromFloat(1000.00).Add(decimal.NewFromFloat(10.00).Mul(decimal.NewFromFloat(numTransfers)))
	expectedBalance3 := decimal.NewFromFloat(1000.00).Sub(decimal.NewFromFloat(20.00).Mul(decimal.NewFromFloat(numTransfers)))
	expectedBalance4 := decimal.NewFromFloat(1000.00).Add(decimal.NewFromFloat(20.00).Mul(decimal.NewFromFloat(numTransfers)))

	assert.True(t, expectedBalance1.Equal(account1.Balance), "expected %v but got %v", expectedBalance1, account1.Balance)
	assert.True(t, expectedBalance2.Equal(account2.Balance), "expected %v but got %v", expectedBalance2, account2.Balance)
	assert.True(t, expectedBalance3.Equal(account3.Balance), "expected %v but got %v", expectedBalance3, account3.Balance)
	assert.True(t, expectedBalance4.Equal(account4.Balance), "expected %v but got %v", expectedBalance4, account4.Balance)
}

func TestConcurrentOppositeTransfers(t *testing.T) {
	svr := setupTestServer()
	defer svr.DB.Migrator().DropTable(&model.Account{}, &model.Transfer{})

	// Create initial accounts
	svr.DB.Create(&model.Account{ID: 1, Balance: decimal.NewFromFloat(1000.00)})
	svr.DB.Create(&model.Account{ID: 2, Balance: decimal.NewFromFloat(1000.00)})

	payload1 := `{"source_account_id": 1, "destination_account_id": 2, "amount": "10.00"}`
	payload2 := `{"source_account_id": 2, "destination_account_id": 1, "amount": "10.00"}`

	transferFunc := func(payload string) {
		req := httptest.NewRequest("POST", "/transactions", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		resp, err := svr.FiberApp.Test(req, 5000)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	}

	const numTransfers = 5
	var wg sync.WaitGroup
	for i := 0; i < numTransfers; i++ {
		wg.Add(1)
		go func(payload string) {
			defer wg.Done()
			transferFunc(payload)
			t.Log("done transferring", "transfer ID", i)
		}(payload1)
		wg.Add(1)
		go func(payload string) {
			defer wg.Done()
			transferFunc(payload)
			t.Log("done transferring", "transfer ID", i)
		}(payload2)
	}
	wg.Wait()

	// Verify the final balances
	var sourceAccount, destinationAccount model.Account
	svr.DB.First(&sourceAccount, 1)
	svr.DB.First(&destinationAccount, 2)

	assert.True(t, decimal.NewFromFloat(1000.00).Equal(sourceAccount.Balance))
	assert.True(t, decimal.NewFromFloat(1000.00).Equal(destinationAccount.Balance))
}

func TestConcurrentReadAndWrite(t *testing.T) {
	svr := setupTestServer()
	defer svr.DB.Migrator().DropTable(&model.Account{}, &model.Transfer{})

	// Create initial accounts
	svr.DB.Create(&model.Account{ID: 1, Balance: decimal.NewFromFloat(1000.00)})
	svr.DB.Create(&model.Account{ID: 2, Balance: decimal.NewFromFloat(1000.00)})

	payload := `{"source_account_id": 1, "destination_account_id": 2, "amount": "10.00"}`

	transferFunc := func() {
		req := httptest.NewRequest("POST", "/transactions", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		resp, err := svr.FiberApp.Test(req, 5000)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	}

	readFunc := func() {
		req := httptest.NewRequest("GET", "/accounts/1", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := svr.FiberApp.Test(req, 5000)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	}

	const numTransfers = 10
	var wg sync.WaitGroup
	for i := 0; i < numTransfers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			transferFunc()
			t.Log("done transferring", "transfer ID", i)
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			readFunc()
			t.Log("done reading", "read ID", i)
		}()
	}
	wg.Wait()

	// Verify the final balances
	var sourceAccount, destinationAccount model.Account
	svr.DB.First(&sourceAccount, 1)
	svr.DB.First(&destinationAccount, 2)

	expectedSourceBalance := decimal.NewFromFloat(1000.00).Sub(decimal.NewFromFloat(10.00).Mul(decimal.NewFromFloat(numTransfers)))
	expectedDestinationBalance := decimal.NewFromFloat(1000.00).Add(decimal.NewFromFloat(10.00).Mul(decimal.NewFromFloat(numTransfers)))

	assert.True(t, expectedSourceBalance.Equal(sourceAccount.Balance), "expected %v but got %v", expectedSourceBalance.String(), sourceAccount.Balance.String())
	assert.True(t, expectedDestinationBalance.Equal(destinationAccount.Balance), "expected %v but got %v", expectedDestinationBalance.String(), destinationAccount.Balance.String())
}
