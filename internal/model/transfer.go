package model

import (
	"github.com/shopspring/decimal"
	"time"
)

type Transfer struct {
	ID                   uint64 `gorm:"primaryKey;autoIncrement"`
	CreatedAt            time.Time
	SourceAccountID      uint64          `gorm:"not null"`
	DestinationAccountID uint64          `gorm:"not null"`
	Amount               decimal.Decimal `gorm:"type:decimal(78,18);not null"`
	SourceAccount        *Account        `gorm:"foreignKey:SourceAccountID"`
	DestinationAccount   *Account        `gorm:"foreignKey:DestinationAccountID"`
}
