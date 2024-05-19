package main

import (
	"fmt"
	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	"internal-transfers-system/config"
	"internal-transfers-system/internal/database"
	"io"
	"log"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"internal-transfers-system/internal/apiserver"
	"internal-transfers-system/internal/model"
)

func setupTestDB() *gorm.DB {
	conf, err := config.LoadConfig("test")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		conf.DBHost,
		conf.DBUser,
		conf.DBPassword,
		conf.DBName,
		conf.DBPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: database.NewLogger(),
	})
	if err != nil {
		log.Fatalf("failed to connect to test database: %v", err)
	}

	if err := db.AutoMigrate(&model.Account{}, &model.Transfer{}); err != nil {
		panic(err)
	}
	return db
}
func setupTestServer() *apiserver.Server {
	app := fiber.New()
	db := setupTestDB()
	svr := apiserver.New(db, app)
	svr.SetupRoutes()
	return svr
}

func TestCreateAccount(t *testing.T) {
	svr := setupTestServer()
	defer svr.DB.Migrator().DropTable(&model.Account{}, &model.Transfer{})

	tests := []struct {
		name       string
		payload    string
		statusCode int
	}{
		{
			name:       "Valid account creation",
			payload:    `{"account_id": 1, "initial_balance": "100.00"}`,
			statusCode: fiber.StatusCreated,
		},
		{
			name:       "Invalid initial balance format",
			payload:    `{"account_id": 2, "initial_balance": "invalid"}`,
			statusCode: fiber.StatusBadRequest,
		},
		{
			name:       "Negative initial balance",
			payload:    `{"account_id": 3, "initial_balance": "-100.00"}`,
			statusCode: fiber.StatusBadRequest,
		},
		{
			name:       "Duplicate account ID",
			payload:    `{"account_id": 1, "initial_balance": "100.00"}`,
			statusCode: fiber.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/accounts", strings.NewReader(tt.payload))
			req.Header.Set("Content-Type", "application/json")

			resp, err := svr.FiberApp.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.statusCode, resp.StatusCode)
		})
	}
}

func TestCreateTransfer(t *testing.T) {
	svr := setupTestServer()
	defer svr.DB.Migrator().DropTable(&model.Account{}, &model.Transfer{})

	// Create initial accounts
	svr.DB.Create(&model.Account{ID: 1, Balance: decimal.NewFromFloat(100.00)})
	svr.DB.Create(&model.Account{ID: 2, Balance: decimal.NewFromFloat(50.00)})

	tests := []struct {
		name       string
		payload    string
		statusCode int
	}{
		{
			name:       "Valid transfer",
			payload:    `{"source_account_id": 1, "destination_account_id": 2, "amount": "50.00"}`,
			statusCode: fiber.StatusCreated,
		},
		{
			name:       "Insufficient funds",
			payload:    `{"source_account_id": 1, "destination_account_id": 2, "amount": "200.00"}`,
			statusCode: fiber.StatusBadRequest,
		},
		{
			name:       "Invalid amount format",
			payload:    `{"source_account_id": 1, "destination_account_id": 2, "amount": "invalid"}`,
			statusCode: fiber.StatusBadRequest,
		},
		{
			name:       "Negative amount",
			payload:    `{"source_account_id": 1, "destination_account_id": 2, "amount": "-1"}`,
			statusCode: fiber.StatusBadRequest,
		},
		{
			name:       "Self transfer",
			payload:    `{"source_account_id": 1, "destination_account_id": 1, "amount": "10.00"}`,
			statusCode: fiber.StatusBadRequest,
		},
		{
			name:       "Missing source account",
			payload:    `{"source_account_id": 3, "destination_account_id": 1, "amount": "10.00"}`,
			statusCode: fiber.StatusNotFound,
		},
		{
			name:       "Missing destination account",
			payload:    `{"source_account_id": 1, "destination_account_id": 3, "amount": "10.00"}`,
			statusCode: fiber.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/transactions", strings.NewReader(tt.payload))
			req.Header.Set("Content-Type", "application/json")

			resp, err := svr.FiberApp.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.statusCode, resp.StatusCode)
		})
	}
}

func TestGetAccount(t *testing.T) {
	svr := setupTestServer()
	defer svr.DB.Migrator().DropTable(&model.Account{}, &model.Transfer{})

	// Create an account
	svr.DB.Create(&model.Account{ID: 1, Balance: decimal.NewFromFloat(100.00)})

	tests := []struct {
		name       string
		accountID  string
		statusCode int
		response   string
	}{
		{
			name:       "Existing account",
			accountID:  "1",
			statusCode: fiber.StatusOK,
			response:   `{"account_id":1,"balance":"100"}`,
		},
		{
			name:       "Non-existent account",
			accountID:  "2",
			statusCode: fiber.StatusNotFound,
			response:   `{"error":"account not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/accounts/%s", tt.accountID)
			req := httptest.NewRequest("GET", url, nil)

			resp, err := svr.FiberApp.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.statusCode, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.JSONEq(t, tt.response, string(body))
		})
	}
}
