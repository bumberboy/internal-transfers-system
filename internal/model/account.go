package model

import (
	"github.com/shopspring/decimal"
	"time"
)

type Account struct {
	ID        uint64 `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Balance   decimal.Decimal `gorm:"type:decimal(78,18);default:0"`
}
