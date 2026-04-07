-- Payment Links
CREATE TABLE payment_links (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id     UUID NOT NULL REFERENCES merchants(id),
    chain           TEXT NOT NULL,
    token           TEXT DEFAULT '',
    amount          TEXT NOT NULL,
    currency        TEXT DEFAULT '',
    description     TEXT DEFAULT '',
    redirect_url    TEXT DEFAULT '',
    is_reusable     BOOLEAN DEFAULT false,
    is_active       BOOLEAN DEFAULT true,
    uses_count      INT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_payment_links_merchant ON payment_links(merchant_id);

-- Audit Log
CREATE TABLE audit_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id     TEXT NOT NULL,
    action          TEXT NOT NULL,
    resource_type   TEXT NOT NULL,
    resource_id     TEXT DEFAULT '',
    ip_address      TEXT DEFAULT '',
    details         JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_audit_logs_merchant ON audit_logs(merchant_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at);

-- Overpayment/Underpayment: add tolerance to payments
ALTER TABLE payments ADD COLUMN IF NOT EXISTS tolerance_percent INT DEFAULT 0;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS payment_status_detail TEXT DEFAULT '';
