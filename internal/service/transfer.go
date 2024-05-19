package service

import (
	"context"
	"errors"
	"github.com/avast/retry-go/v4"
	"github.com/jackc/pgx/v5/pgconn"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"internal-transfers-system/internal/apimodel"
	"internal-transfers-system/internal/model"
	"internal-transfers-system/internal/svrerror"
)

// ProcessTransfer uses optimistic concurrency control by looking at the updatedAt timestamp on the account
// before updating the account values
func ProcessTransfer(ctx context.Context, db *gorm.DB, transfer apimodel.TransferRequest, amount decimal.Decimal) error {
	return retry.Do(
		func() error {
			return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				slog.Debug("processing transfer", "from", transfer.SourceAccountID, "to", transfer.DestinationAccountID)
				var sourceAccount, destinationAccount model.Account

				if err := tx.Take(&sourceAccount, "id = ?", transfer.SourceAccountID).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return svrerror.New("source account not found", http.StatusNotFound)
					}
					return err
				}

				if err := tx.Take(&destinationAccount, "id = ?", transfer.DestinationAccountID).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return svrerror.New("destination account not found", http.StatusNotFound)
					}
					return err
				}

				if sourceAccount.Balance.LessThan(amount) {
					return svrerror.New("insufficient funds", http.StatusBadRequest)
				}

				updatedSourceBalance := sourceAccount.Balance.Sub(amount)
				updatedDestinationBalance := destinationAccount.Balance.Add(amount)

				// Combine the update of both records into a single query as an optimisation
				result := tx.Exec(`
                        UPDATE accounts
                        SET balance = CASE 
                            WHEN id = ? AND updated_at = ? THEN ?
                            WHEN id = ? AND updated_at = ? THEN ?
                            ELSE balance
                        END,
                        updated_at = CASE
                            WHEN id = ? AND updated_at = ? THEN NOW()
                            WHEN id = ? AND updated_at = ? THEN NOW()
                            ELSE updated_at
                        END
                        WHERE (id = ? AND updated_at = ?) 
                           OR (id = ? AND updated_at = ?)`,
					sourceAccount.ID, sourceAccount.UpdatedAt, updatedSourceBalance,
					destinationAccount.ID, destinationAccount.UpdatedAt, updatedDestinationBalance,
					sourceAccount.ID, sourceAccount.UpdatedAt,
					destinationAccount.ID, destinationAccount.UpdatedAt,
					sourceAccount.ID, sourceAccount.UpdatedAt,
					destinationAccount.ID, destinationAccount.UpdatedAt,
				)

				if result.Error != nil {
					return result.Error
				}
				if result.RowsAffected != 2 {
					return svrerror.New("account updatedAt mismatch, retrying", http.StatusConflict)
				}

				newTransfer := model.Transfer{
					SourceAccountID:      transfer.SourceAccountID,
					DestinationAccountID: transfer.DestinationAccountID,
					Amount:               amount,
				}

				if err := tx.Create(&newTransfer).Error; err != nil {
					return err
				}

				return nil
			})
		},
		retry.Attempts(5),
		retry.Delay(100*time.Millisecond),
		retry.MaxDelay(2*time.Second),
		retry.MaxJitter(100*time.Millisecond),
		retry.DelayType(retry.CombineDelay(
			retry.BackOffDelay,
			retry.RandomDelay,
		)),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("retry: #%d: %s\n", n, err)
		}),
		retry.RetryIf(func(err error) bool {
			var svrError *svrerror.Error
			if errors.As(err, &svrError) && svrError.StatusCode == http.StatusConflict {
				slog.Warn("transfer: updatedAt mismatch, retrying")
				return true
			}
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "55P03" {
				slog.Warn("transfer: failed to get a lock")
				return true
			}
			return false
		}),
	)
}
