CREATE TABLE IF NOT EXISTS accounts
(
    id         BIGINT PRIMARY KEY,
    created_at TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    balance    NUMERIC(78, 18) NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS transfers
(
    id                     BIGSERIAL PRIMARY KEY,
    created_at             TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    source_account_id      BIGINT          NOT NULL,
    destination_account_id BIGINT          NOT NULL,
    amount                 NUMERIC(78, 18) NOT NULL,
    CONSTRAINT fk_source_account
        FOREIGN KEY (source_account_id)
            REFERENCES accounts (id),
    CONSTRAINT fk_destination_account
        FOREIGN KEY (destination_account_id)
            REFERENCES accounts (id)
);
