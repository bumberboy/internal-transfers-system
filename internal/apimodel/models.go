package apimodel

// These are models used at the api presentation layer. I've put them in the same file but as the project grows, we can refactor and split them out.

type TransferRequest struct {
	SourceAccountID      uint64 `json:"source_account_id"`
	DestinationAccountID uint64 `json:"destination_account_id"`
	Amount               string `json:"amount"`
}

type CreateAccountRequest struct {
	AccountID      uint64 `json:"account_id"`
	InitialBalance string `json:"initial_balance"`
}
