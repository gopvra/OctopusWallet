-- ============================================================
-- Enterprise Features Migration
-- ============================================================

-- Feature 1: Auto-Sweep
CREATE TABLE merchant_collection_addresses (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id     UUID NOT NULL REFERENCES merchants(id),
    chain           TEXT NOT NULL,
    address         TEXT NOT NULL,
    sweep_threshold TEXT NOT NULL DEFAULT '0',
    is_active       BOOLEAN DEFAULT true,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE(merchant_id, chain)
);

CREATE TABLE sweep_tasks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id     UUID NOT NULL REFERENCES merchants(id),
    payment_id      UUID REFERENCES payments(id),
    chain           TEXT NOT NULL,
    token           TEXT DEFAULT '',
    from_address    TEXT NOT NULL,
    to_address      TEXT NOT NULL,
    amount          TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    tx_hash         TEXT,
    gas_deposit_id  UUID,
    error_message   TEXT,
    retry_count     INT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_sweep_tasks_status ON sweep_tasks(status);

-- Feature 2: Cold/Hot Wallet
CREATE TABLE cold_wallet_configs (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id             UUID NOT NULL REFERENCES merchants(id),
    chain                   TEXT NOT NULL,
    cold_wallet_address     TEXT NOT NULL DEFAULT '',
    hot_wallet_max_balance  TEXT NOT NULL DEFAULT '0',
    enabled                 BOOLEAN DEFAULT false,
    created_at              TIMESTAMPTZ DEFAULT now(),
    updated_at              TIMESTAMPTZ DEFAULT now(),
    UNIQUE(merchant_id, chain)
);

CREATE TABLE wallet_transfers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id     UUID NOT NULL REFERENCES merchants(id),
    chain           TEXT NOT NULL,
    token           TEXT DEFAULT '',
    from_address    TEXT NOT NULL,
    to_address      TEXT NOT NULL,
    amount          TEXT NOT NULL,
    transfer_type   TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    tx_hash         TEXT,
    error_message   TEXT,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_wallet_transfers_status ON wallet_transfers(status);

-- Feature 3: Withdrawal Approval
CREATE TABLE approval_configs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id         UUID NOT NULL REFERENCES merchants(id) UNIQUE,
    approval_threshold  TEXT NOT NULL DEFAULT '0',
    single_tx_limit     TEXT NOT NULL DEFAULT '0',
    daily_limit         TEXT NOT NULL DEFAULT '0',
    auto_release        BOOLEAN DEFAULT false,
    enabled             BOOLEAN DEFAULT true,
    created_at          TIMESTAMPTZ DEFAULT now(),
    updated_at          TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE payout_approvals (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payout_id       UUID NOT NULL REFERENCES payouts(id),
    merchant_id     UUID NOT NULL REFERENCES merchants(id),
    action          TEXT NOT NULL,
    approver_id     TEXT NOT NULL,
    approver_note   TEXT DEFAULT '',
    created_at      TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_payout_approvals_payout ON payout_approvals(payout_id);

ALTER TABLE payouts ADD COLUMN approval_status TEXT DEFAULT '';

CREATE TABLE payout_daily_totals (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id     UUID NOT NULL REFERENCES merchants(id),
    chain           TEXT NOT NULL,
    date            DATE NOT NULL,
    total_amount    TEXT NOT NULL DEFAULT '0',
    updated_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE(merchant_id, chain, date)
);

-- Feature 4: Gas Fee Management
CREATE TABLE gas_deposits (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain           TEXT NOT NULL,
    to_address      TEXT NOT NULL,
    amount          TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    tx_hash         TEXT,
    sweep_task_id   UUID REFERENCES sweep_tasks(id),
    error_message   TEXT,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_gas_deposits_status ON gas_deposits(status);

CREATE TABLE gas_alerts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain           TEXT NOT NULL,
    station_address TEXT NOT NULL,
    balance         TEXT NOT NULL,
    threshold       TEXT NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT now()
);
