CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE merchants (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT NOT NULL,
    email         TEXT NOT NULL UNIQUE,
    api_key_hash  TEXT NOT NULL UNIQUE,
    webhook_url   TEXT DEFAULT '',
    is_active     BOOLEAN DEFAULT true,
    created_at    TIMESTAMPTZ DEFAULT now(),
    updated_at    TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE wallets (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id      UUID NOT NULL REFERENCES merchants(id),
    chain            TEXT NOT NULL,
    address          TEXT NOT NULL,
    derivation_index INT NOT NULL,
    created_at       TIMESTAMPTZ DEFAULT now(),
    UNIQUE(chain, address)
);
CREATE INDEX idx_wallets_merchant ON wallets(merchant_id);
CREATE INDEX idx_wallets_address ON wallets(address);

CREATE TABLE payments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id     UUID NOT NULL REFERENCES merchants(id),
    chain           TEXT NOT NULL,
    token           TEXT DEFAULT '',
    amount_expected TEXT NOT NULL,
    amount_received TEXT DEFAULT '0',
    address         TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    tx_hash         TEXT,
    confirmations   INT DEFAULT 0,
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_payments_address ON payments(address);
CREATE INDEX idx_payments_status ON payments(status);

CREATE TABLE payouts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id   UUID NOT NULL REFERENCES merchants(id),
    chain         TEXT NOT NULL,
    token         TEXT DEFAULT '',
    to_address    TEXT NOT NULL,
    amount        TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'pending',
    tx_hash       TEXT,
    error_message TEXT,
    created_at    TIMESTAMPTZ DEFAULT now(),
    updated_at    TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE chain_state (
    chain              TEXT PRIMARY KEY,
    last_scanned_block BIGINT NOT NULL DEFAULT 0,
    updated_at         TIMESTAMPTZ DEFAULT now()
);
