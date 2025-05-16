CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(100) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT (now() AT TIME ZONE 'utc') NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT (now() AT TIME ZONE 'utc') NOT NULL
);
COMMENT ON COLUMN users.id IS 'Unique user ID (auto-incremented)';
COMMENT ON COLUMN users.username IS 'Unique username for the user';
COMMENT ON COLUMN users.password IS 'User''s password (plaintext or hashed)';
COMMENT ON COLUMN users.active IS 'Flag indicating whether the user is active (default is TRUE)';
COMMENT ON COLUMN users.created_at IS 'Timestamp when the user was created (defaults to current UTC time)';
COMMENT ON COLUMN users.updated_at IS 'Timestamp when the user was last updated (defaults to current UTC time)';


CREATE TABLE IF NOT EXISTS wallets (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    currency CHAR(3) NOT NULL,
    amount INTEGER NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT (now() AT TIME ZONE 'utc') NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT (now() AT TIME ZONE 'utc') NOT NULL
);
CREATE UNIQUE INDEX uk_wallet_id ON wallets(username, currency);

COMMENT ON COLUMN wallets.id IS 'Unique wallet ID (auto-incremented)';
COMMENT ON COLUMN wallets.username IS 'Username of the user who owns this wallet';
COMMENT ON COLUMN wallets.amount IS 'Wallet balance stored as a integer (smallest unit of currency)';
COMMENT ON COLUMN wallets.currency IS 'Currency type (e.g., SGD, USD, EUR)';
COMMENT ON COLUMN wallets.status IS 'The current status of the wallet, can be 'active' or 'frozen' ';
COMMENT ON COLUMN wallets.created_at IS 'Timestamp when the wallet was created (defaults to current UTC time)';
COMMENT ON COLUMN wallets.updated_at IS 'Timestamp when the wallet was last updated (defaults to current UTC time)';


CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    uid CHAR(20) NOT NULL,
    type VARCHAR(16) NOT NULL, 
    initiated_by VARCHAR(100) NOT NULL DEFAULT '',
    currency CHAR(3) NOT NULL,
    amount INTEGER NOT NULL,
    status VARCHAR(16) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT (now() AT TIME ZONE 'utc') NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT (now() AT TIME ZONE 'utc') NOT NULL
);
CREATE UNIQUE INDEX uk_uid ON transactions(uid);

COMMENT ON COLUMN transactions.id IS 'Unique transaction ID (auto-incremented)';
COMMENT ON COLUMN transactions.uid IS 'Unique transaction ID (32 character UUID without hyphens)';
COMMENT ON COLUMN transactions.type IS 'The type of transaction (e.g., ''deposit'', ''withdrawal'', ''transfer'')';
COMMENT ON COLUMN transactions.initiated_by IS 'The account initiating the transaction (It can be a wallet ID, or internal amdmin)';
COMMENT ON COLUMN transactions.currency IS 'The currency used in the transaction (e.g., SGD, JPY and etc)';
COMMENT ON COLUMN transactions.amount IS 'The amount involved in the transaction (integer, smallest unit of currency)';
COMMENT ON COLUMN transactions.status IS 'The current status of the transaction (e.g., ''pending'', ''completed'', ''failed'')';
COMMENT ON COLUMN transactions.metadata IS 'Optional key-value pair to store additional information related to this transactions';
COMMENT ON COLUMN transactions.created_at IS 'Timestamp when the transaction was created (defaults to current UTC time)';
COMMENT ON COLUMN transactions.updated_at IS 'Timestamp when the transaction was last updated (defaults to current UTC time)';


CREATE TABLE ledgers (
    id SERIAL PRIMARY KEY,
    tx_uid CHAR(20) NOT NULL,
    username VARCHAR(100) NOT NULL,
    currency CHAR(3) NOT NULL,
    amount INTEGER NOT NULL,
    direction CHAR(1) NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT (now() AT TIME ZONE 'utc') NOT NULL
);

CREATE INDEX idx_wallet ON ledgers(username, currency, created_at DESC);
CREATE INDEX idx_tx_uid ON ledgers(tx_uid);

COMMENT ON COLUMN ledgers.id IS 'Unique ledger entry ID (auto-incremented)';
COMMENT ON COLUMN ledgers.tx_uid IS 'Unique transaction ID this ledger entry is associated with';
COMMENT ON COLUMN ledgers.username IS 'Username of the user whose wallet is affected by this ledger entry';
COMMENT ON COLUMN ledgers.currency IS 'Currency of the amount in this ledger entry';
COMMENT ON COLUMN ledgers.direction IS 'Indicates if the amount is a credit (''+'') or debit (''-'')';
COMMENT ON COLUMN ledgers.amount IS 'The amount of the transaction affecting the wallet balance (integer, smallest unit of currency)';
COMMENT ON COLUMN ledgers.created_at IS 'Timestamp when the ledger entry was created';

