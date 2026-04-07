-- ============================================================
-- Feature: Invoice System (enhanced payments with metadata)
-- ============================================================

ALTER TABLE payments ADD COLUMN currency TEXT DEFAULT '';
ALTER TABLE payments ADD COLUMN description TEXT DEFAULT '';
ALTER TABLE payments ADD COLUMN metadata JSONB DEFAULT '{}';
ALTER TABLE payments ADD COLUMN redirect_url TEXT DEFAULT '';
ALTER TABLE payments ADD COLUMN order_id TEXT DEFAULT '';
CREATE INDEX idx_payments_order_id ON payments(order_id) WHERE order_id != '';

-- ============================================================
-- Feature: Refund Support
-- ============================================================

CREATE TABLE refunds (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id      UUID NOT NULL REFERENCES payments(id),
    merchant_id     UUID NOT NULL REFERENCES merchants(id),
    chain           TEXT NOT NULL,
    token           TEXT DEFAULT '',
    to_address      TEXT NOT NULL,
    amount          TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',  -- pending, processing, completed, failed
    tx_hash         TEXT,
    reason          TEXT DEFAULT '',
    error_message   TEXT,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_refunds_payment ON refunds(payment_id);
CREATE INDEX idx_refunds_status ON refunds(status);

-- ============================================================
-- Feature: Merchant Ledger / Balance Tracking
-- ============================================================

CREATE TABLE merchant_balances (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id     UUID NOT NULL REFERENCES merchants(id),
    chain           TEXT NOT NULL,
    token           TEXT DEFAULT '',
    available       TEXT NOT NULL DEFAULT '0',
    pending         TEXT NOT NULL DEFAULT '0',
    updated_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE(merchant_id, chain, token)
);
CREATE INDEX idx_merchant_balances_merchant ON merchant_balances(merchant_id);

-- ============================================================
-- Feature: Supported Currencies Registry
-- ============================================================

CREATE TABLE supported_currencies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain           TEXT NOT NULL,
    symbol          TEXT NOT NULL,
    name            TEXT NOT NULL,
    token_address   TEXT DEFAULT '',
    decimals        INT NOT NULL DEFAULT 18,
    is_native       BOOLEAN DEFAULT false,
    is_active       BOOLEAN DEFAULT true,
    min_amount      TEXT DEFAULT '0',
    created_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE(chain, symbol)
);

-- Seed default native currencies
INSERT INTO supported_currencies (chain, symbol, name, is_native, decimals) VALUES
    ('ethereum', 'ETH', 'Ethereum', true, 18),
    ('bsc', 'BNB', 'BNB', true, 18),
    ('polygon', 'MATIC', 'Polygon', true, 18),
    ('solana', 'SOL', 'Solana', true, 9),
    ('tron', 'TRX', 'TRON', true, 6),
    ('bitcoin', 'BTC', 'Bitcoin', true, 8);

-- ============================================================
-- Feature: Batch Payouts
-- ============================================================

CREATE TABLE batch_payouts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id     UUID NOT NULL REFERENCES merchants(id),
    chain           TEXT NOT NULL,
    token           TEXT DEFAULT '',
    total_amount    TEXT NOT NULL DEFAULT '0',
    total_count     INT NOT NULL DEFAULT 0,
    completed_count INT NOT NULL DEFAULT 0,
    failed_count    INT NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'pending',  -- pending, processing, completed, partial, failed
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE batch_payout_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id        UUID NOT NULL REFERENCES batch_payouts(id),
    payout_id       UUID REFERENCES payouts(id),
    to_address      TEXT NOT NULL,
    amount          TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    error_message   TEXT,
    created_at      TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_batch_payout_items_batch ON batch_payout_items(batch_id);

-- ============================================================
-- Feature: IP Whitelist Storage
-- ============================================================

CREATE TABLE merchant_ip_whitelist (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id     UUID NOT NULL REFERENCES merchants(id),
    ip_address      TEXT NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE(merchant_id, ip_address)
);
CREATE INDEX idx_ip_whitelist_merchant ON merchant_ip_whitelist(merchant_id);
