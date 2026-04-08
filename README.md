# OctopusWallet

Open-source multi-chain merchant payment gateway. Self-hosted alternative to BitPay / CoinsPaid with enterprise features including auto-sweep, cold/hot wallet separation, withdrawal approval workflows, gas fee management, payment links, audit logging, RBAC admin panel, and Redis-backed caching.

## Hosted Checkout Page

<p align="center">
  <img src="docs/checkout-pending.svg" alt="Waiting for Payment" width="260" />
  <img src="docs/checkout-confirming.svg" alt="Confirming" width="260" />
  <img src="docs/checkout-completed.svg" alt="Payment Complete" width="260" />
</p>

Customers see a hosted payment page at `/pay/:id` with QR code scanning, one-click address copy, countdown timer, and real-time status updates via WebSocket.

## Supported Chains

| Chain | Native | Token Standard | Address Format |
|-------|--------|----------------|----------------|
| Ethereum | ETH | ERC-20 | 0x... |
| BSC | BNB | BEP-20 | 0x... |
| Polygon | MATIC | ERC-20 | 0x... |
| Solana | SOL | SPL Token | Base58 |
| TRON | TRX | TRC-20 | Base58Check |
| Bitcoin | BTC | - | Bech32 (Segwit) |

## Architecture

```
                    ┌──────────────┐
  Merchant API ───> │  API Server  │ ──── PostgreSQL (via GORM)
  Dashboard UI ───> │  (+ Web UI)  │ ──── Redis (cache + rate limit)
                    └──────────────┘
                    ┌──────────────┐
  Blockchains ───> │   Worker     │ ──── PostgreSQL (via GORM)
                    │  - Monitor   │
                    │  - Payout    │
                    │  - Refund    │
                    │  - Sweep     │
                    │  - GasStation│
                    │  - ColdWallet│
                    └──────────────┘
                          │
                    Webhook ──────> Merchant
```

**Data layer:** All database access uses [GORM](https://gorm.io/) (`gorm.io/gorm` + `gorm.io/driver/postgres`). The previous sqlx/raw-SQL layer has been fully replaced — all 108 queries are now expressed through the GORM query builder.

**Caching layer:** [go-redis/v9](https://github.com/redis/go-redis) provides:
- **Rate limiting** — sliding-window algorithm backed by Redis sorted sets
- **Idempotency store** — `SET NX` with TTL for deduplication
- **Graceful degradation** — falls back to in-memory stores when Redis is unavailable

### Worker Services

| Service | Description |
|---------|-------------|
| **Monitor** | Watches blockchains for incoming payments, tracks confirmations, triggers auto-sweep on completion |
| **Payout** | Processes approved payouts — signs and broadcasts withdrawal transactions |
| **Refund** | Processes pending refunds — derives keys, sends refund tx, updates merchant balance |
| **Sweep** | Collects funds from hot wallets to designated collection addresses |
| **GasStation** | Manages gas/fee balances for sweep and payout transactions |
| **ColdWallet** | Transfers excess hot wallet funds to cold storage based on configured thresholds |

## Quick Start

```bash
# 1. Start PostgreSQL + Redis
docker-compose up -d postgres redis

# 2. Configure
cp config/config.example.yaml config/config.yaml
# Edit config.yaml — set wallet seed, database URL, redis.addr, admin.jwt_secret

# 3. Run migrations
DATABASE_URL="postgres://octopus:octopus@localhost:5432/octopus_wallet?sslmode=disable" make migrate

# 4. Run backend
make run-server   # API on :8080
make run-worker   # Background services

# 5. Run frontend (development)
make web-install  # Install npm dependencies
make web-dev      # Vite dev server with HMR
```

> **Note:** Redis is optional. If Redis is unreachable the server starts normally and falls back to in-memory rate limiting and idempotency stores.

### Docker (full stack)

```bash
docker-compose up -d   # Starts postgres + redis + server + worker
```

The Dockerfile includes a Node.js build stage that compiles the React frontend and bundles it into the server image.

### Makefile Commands

| Command | Description |
|---------|-------------|
| `make build` | Build server + worker binaries |
| `make run-server` | Run API server |
| `make run-worker` | Run background worker |
| `make test` | Run all Go tests |
| `make migrate` | Run all SQL migrations |
| `make web-install` | Install frontend npm dependencies |
| `make web-dev` | Start Vite dev server |
| `make web-build` | Build frontend for production |
| `make docker-up` | Start all services via docker-compose |
| `make docker-down` | Stop all services |
| `make clean` | Remove build artifacts |

## Web Dashboard

The included React dashboard provides a merchant management UI.

| Route | Page |
|-------|------|
| `/login` | Merchant login (API key) |
| `/dashboard` | Overview — balances, recent activity |
| `/dashboard/payments` | Payment list with status tracking |
| `/dashboard/payouts` | Payout list + create / approve / reject |
| `/dashboard/refunds` | Refund management — create + search by payment |
| `/dashboard/sweeps` | Auto-sweep config + task history |
| `/dashboard/settings` | Approval rules, IP whitelist |
| `/pay/:id` | Customer-facing checkout page (with real-time WebSocket updates) |

## API Reference

### Unified Response Format

All API responses use a consistent envelope:

```json
// Success
{"code": 0, "msg": "ok", "data": { ... }}

// Error
{"code": 20001, "msg": "payment not found"}
```

**Error code ranges:**

| Range | Domain |
|-------|--------|
| `1xxxx` | General (auth, validation, rate limit) |
| `2xxxx` | Payment |
| `3xxxx` | Payout |
| `4xxxx` | Refund |
| `5xxxx` | Merchant |
| `6xxxx` | Admin |

**i18n error messages:** Error messages are returned in the caller's language. Set the `Accept-Language` header or append `?lang=` query parameter. Supported: `en`, `zh`, `ja`, `ko`, `es`.

### Public Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check + active chains |
| POST | `/api/v1/merchants/register` | Register merchant, get API key |
| GET | `/api/v1/currencies` | List supported currencies |
| GET | `/api/v1/rates?chain=ethereum` | Get fee estimates per chain |
| GET | `/api/v1/payment-links/:id` | Get payment link details (for checkout) |

### Payment / Invoice

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/payments/create` | Create payment invoice |
| GET | `/api/v1/payments/:id` | Get payment status |
| GET | `/api/v1/payments?limit=20&offset=0` | List payments (paginated) |

**Create Payment** accepts: `chain`, `amount`, `token`, `currency`, `description`, `order_id`, `redirect_url`

### Refunds

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/refunds/create` | Refund a completed payment |
| GET | `/api/v1/refunds/:id` | Get refund status |
| GET | `/api/v1/payments/:id/refunds` | List refunds for a payment |

Refund amount is validated against `amount_received` — the sum of all refunds for a payment cannot exceed the original payment amount.

### Payouts (with Approval Workflow)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/payouts/create` | Create payout (subject to approval rules) |
| GET | `/api/v1/payouts/:id` | Get payout status |
| GET | `/api/v1/payouts?limit=20&offset=0` | List payouts (paginated) |
| POST | `/api/v1/payouts/:id/approve` | Approve pending payout |
| POST | `/api/v1/payouts/:id/reject` | Reject pending payout |

### Batch Payouts (Mass Payouts)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/payouts/batch` | Create batch payout (up to 100 items) |
| GET | `/api/v1/payouts/batch/:id` | Get batch status + items |
| GET | `/api/v1/payouts/batches` | List batch payouts |

Each batch item is individually checked against approval limits and daily payout caps.

### Payment Links

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/payment-links` | Create shareable payment link |
| GET | `/api/v1/payment-links` | List merchant's payment links |

Payment links can be **reusable** (multiple payments) or **single-use**. Accepts: `chain`, `amount`, `token`, `currency`, `description`, `redirect_url`, `is_reusable`.

### Approval Configuration

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/approval/config` | Set approval rules |
| GET | `/api/v1/approval/config` | Get approval rules |

Configurable: `approval_threshold`, `single_tx_limit`, `daily_limit`, `auto_release`

### Balance / Ledger

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/balances` | Merchant balance per chain/token |
| GET | `/api/v1/wallets` | List derived wallet addresses |

Balances are automatically updated: +amount on payment completion, -amount on payout/refund completion.

### Auto-Sweep (Fund Collection)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/sweep/collection-address` | Set collection address per chain |
| GET | `/api/v1/sweep/collection-address` | List collection addresses |
| GET | `/api/v1/sweep/tasks` | List sweep tasks |

### Cold/Hot Wallet

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/cold-wallet/config` | Configure cold wallet + threshold |
| GET | `/api/v1/cold-wallet/config` | Get cold wallet configs |
| GET | `/api/v1/cold-wallet/transfers` | List hot/cold transfers |

### Gas Station

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/gas/status` | Gas station balances per chain |

### Export (CSV / JSON)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/export/payments?format=csv&from=&to=` | Export payments |
| GET | `/api/v1/export/payouts?format=csv&from=&to=` | Export payouts |

Supports `format=csv` (default) or `format=json`. Date range filtering via `from` and `to` query params (ISO 8601).

### Audit Logs

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/audit-logs?limit=50&offset=0` | List audit logs |

All mutating API calls (POST/PUT/DELETE) are automatically recorded with merchant ID, IP address, path, and timestamp.

### Security

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/security/ip-whitelist` | Set IP whitelist |
| GET | `/api/v1/security/ip-whitelist` | Get IP whitelist |

### WebSocket

| Endpoint | Description |
|----------|-------------|
| `GET /ws/payments/:id` | Real-time payment status updates |

WebSocket connections enforce origin validation, per-IP connection limits, and UUID format validation on the payment ID parameter.

## Payment Flow

```
1. Merchant: POST /payments/create {chain, amount, description, order_id}
2. System:   Derives fresh HD address → returns {id, address, amount, expires_at}
3. Customer: Sends crypto to the address
4. Worker:   Detects tx on-chain → status: confirming → Webhook: payment.confirming
5. Worker:   Confirmations met → status: completed → Webhook: payment.completed
6. Worker:   Balance updated → Auto-sweep to collection address (if configured)
```

## Payout Flow (Approval + Auto/Manual Release)

```
1. Merchant: POST /payouts/create {chain, to_address, amount}
2. System:   Check single_tx_limit → Check daily_limit → Determine approval:
             ├── auto_release=true AND amount < threshold → auto-release
             └── otherwise → status: pending_approval → Webhook: payout.pending_approval
3. System:   Daily payout total incremented for limit enforcement
4. Approver: POST /payouts/:id/approve → Webhook: payout.approved
5. Worker:   Signs + broadcasts transaction → Balance deducted → Webhook: payout.completed
```

## Webhook Events

| Event | Description |
|-------|-------------|
| `payment.confirming` | Transaction detected, awaiting confirmations |
| `payment.completed` | Required confirmations reached |
| `payment.expired` | Payment expired (30 min) |
| `payout.pending_approval` | Payout awaiting manual approval |
| `payout.approved` | Payout approved |
| `payout.rejected` | Payout rejected |
| `payout.completed` | Payout transaction confirmed |
| `payout.failed` | Payout failed |
| `refund.completed` | Refund transaction confirmed |
| `refund.failed` | Refund failed |
| `sweep.completed` | Auto-sweep completed |
| `sweep.failed` | Auto-sweep failed |
| `transfer.completed` | Hot/cold transfer completed |
| `transfer.failed` | Hot/cold transfer failed |
| `gas.deposit_completed` | Gas deposited for sweep |
| `gas.low_balance` | Gas station low balance alert |

All webhooks include `X-Webhook-Signature` (HMAC-SHA256) header for verification. Webhook delivery URLs are validated against SSRF (private/internal IPs are rejected).

## Security Features

| Feature | Description |
|---------|-------------|
| **API Key Auth** | SHA-256 hashed keys, never stored in plaintext |
| **Request Signing** | Optional HMAC-SHA256 on requests (`X-Request-Signature`) |
| **Webhook Signing** | HMAC-SHA256 on all webhook payloads |
| **Idempotency** | `X-Idempotency-Key` header prevents duplicate requests (Redis `SET NX` with TTL; in-memory fallback) |
| **IP Whitelist** | Per-merchant IP restriction, enforced in middleware |
| **Rate Limiting** | Sliding-window rate limiter backed by Redis sorted sets (in-memory fallback when Redis is unavailable) |
| **RBAC (Admin)** | 3 roles (`super_admin`, `admin`, `viewer`) with 23 granular permissions; every admin endpoint gated with `RequirePermission()` |
| **SSRF Prevention** | Webhook URLs validated to reject private/loopback/internal addresses |
| **ILIKE Escaping** | User-supplied search strings are escaped before use in SQL `ILIKE` patterns to prevent wildcard injection |
| **Refund Validation** | Sum of all refund amounts for a payment cannot exceed `amount_received` |
| **Approval Threshold Bypass Prevention** | Approval thresholds enforced server-side; requests cannot bypass limits via direct API calls |
| **HSTS + CSP Headers** | Strict-Transport-Security and Content-Security-Policy headers set on all responses |
| **JWT Secret Required** | Server refuses to start if `admin.jwt_secret` is missing or empty |
| **Random Password Generation** | If `admin.default_pass` is empty, a secure random password is generated and logged at startup |
| **CORS Custom Headers** | CORS configuration exposes API-key and signature headers for cross-origin clients |
| **WebSocket Hardening** | Origin validation, per-IP connection limits, UUID format validation on payment IDs |
| **Private Key Zeroing** | Key material wiped from memory after use |
| **HD Wallet** | BIP-39/32/44 deterministic address derivation |
| **Cold/Hot Separation** | Auto-transfer excess funds to cold storage |
| **Approval Workflow** | Configurable thresholds + daily/single-tx limits |
| **Atomic Processing** | SELECT FOR UPDATE SKIP LOCKED prevents double-processing |
| **Input Validation** | Amount (positive integer), address format (per-chain regex) |
| **Audit Log** | All mutating API calls recorded with IP + timestamp |
| **Session Storage** | API keys stored in sessionStorage (cleared on browser close) |

## Feature Comparison

| Feature | BitPay | CoinsPaid | OctopusWallet |
|---------|--------|-----------|---------------|
| Multi-chain | Limited | 20+ | 6 chains |
| Payment/Invoice | ✓ | ✓ | ✓ |
| Payment Links | ✓ | ✓ | ✓ |
| Refunds | ✓ | ✓ | ✓ |
| Batch Payouts | ✓ | ✓ (CSV) | ✓ (API, up to 100) |
| Approval Workflow | ✓ | ✓ | ✓ |
| Auto-Sweep | - | - | ✓ |
| Cold/Hot Wallet | ✓ | ✓ | ✓ |
| Gas Fee Management | - | - | ✓ |
| CSV/JSON Export | ✓ | ✓ | ✓ |
| Audit Log | ✓ | ✓ | ✓ |
| Webhook HMAC | Custom | HMAC | SHA-256 HMAC |
| Idempotency | ✓ | ✓ | ✓ (Redis-backed) |
| IP Whitelist | - | ✓ | ✓ |
| Request Signing | ✓ | ✓ | ✓ |
| Rate Limiting | - | - | ✓ (Redis sliding window) |
| Supported Currencies API | ✓ | ✓ | ✓ |
| Balance/Ledger | ✓ | ✓ | ✓ |
| Pagination | ✓ | ✓ | ✓ |
| Unified Error Codes | - | - | ✓ (6 domains, 5-lang i18n) |
| RBAC Admin | ✓ | ✓ | ✓ (3 roles, 23 permissions) |
| Admin Dashboard | ✓ | ✓ | ✓ (React, dark-theme) |
| Real-time WebSocket | - | - | ✓ |
| SSRF Protection | - | - | ✓ |
| Self-hosted | - | - | ✓ |
| Open Source | - | - | ✓ (Apache 2.0) |

## Configuration

Set via `config/config.yaml` or environment variables (prefix `OCTOPUS_`):

| Config | Env Var | Description |
|--------|---------|-------------|
| `wallet.master_seed` | `OCTOPUS_WALLET_MASTER_SEED` | BIP-39 mnemonic |
| `wallet.encryption_key` | `OCTOPUS_WALLET_ENCRYPTION_KEY` | 32-byte hex AES key |
| `database.url` | `OCTOPUS_DATABASE_URL` | PostgreSQL connection |
| `redis.addr` | `OCTOPUS_REDIS_ADDR` | Redis address (e.g. `localhost:6379`) |
| `redis.password` | `OCTOPUS_REDIS_PASSWORD` | Redis password (optional) |
| `redis.db` | `OCTOPUS_REDIS_DB` | Redis database number (default `0`) |
| `chains.<name>.rpc_url` | - | Chain RPC endpoint |
| `gas_station.enabled` | - | Enable gas fee management |
| `gas_station.chains.<name>.station_address` | - | Gas station address |

> **Redis is optional.** When `redis.addr` is empty or Redis is unreachable, the server automatically falls back to in-memory rate limiting and idempotency stores. For production deployments Redis is recommended.

## Admin Dashboard

OctopusWallet ships with a full-featured admin panel for platform operators. The admin API runs alongside the merchant API — no additional services required.

**Repository**: [OctopusWallet-Admin](https://github.com/gopvra/OctopusWallet-Admin)

<p align="center">
  <a href="https://github.com/gopvra/OctopusWallet-Admin">
    <img src="https://raw.githubusercontent.com/gopvra/OctopusWallet-Admin/claude/octopus-admin-system-sFukw/docs/dashboard-preview.svg" alt="Admin Dashboard" width="80%" />
  </a>
</p>

<p align="center">
  <a href="https://github.com/gopvra/OctopusWallet-Admin">
    <img src="https://raw.githubusercontent.com/gopvra/OctopusWallet-Admin/claude/octopus-admin-system-sFukw/docs/login-preview.svg" alt="Admin Login" width="45%" />
  </a>
  <a href="https://github.com/gopvra/OctopusWallet-Admin">
    <img src="https://raw.githubusercontent.com/gopvra/OctopusWallet-Admin/claude/octopus-admin-system-sFukw/docs/architecture.svg" alt="Architecture" width="45%" />
  </a>
</p>

### RBAC Roles & Permissions

The admin panel uses role-based access control with three built-in roles:

| Role | Access Level |
|------|-------------|
| `super_admin` | Full access — all 23 permissions, can manage admin users and roles |
| `admin` | Read-only access across all resources |
| `viewer` | Dashboard view only |

**23 granular permissions** are enforced individually on every admin endpoint via `RequirePermission()` middleware:

`dashboard:view`, `merchant:list`, `merchant:view`, `merchant:update`, `merchant:toggle`, `payment:list`, `payment:view`, `payout:list`, `payout:view`, `refund:list`, `refund:view`, `batch:list`, `batch:view`, `wallet:list`, `wallet:view`, `balance:list`, `balance:view`, `currency:list`, `currency:view`, `chain:view`, `admin:list`, `admin:create`, `admin:delete`

### Admin Capabilities

| Category | Features |
|----------|----------|
| **Dashboard** | Real-time stats, payment volume charts, chain distribution, recent activity feed |
| **Merchants** | List, search, detail view, activate/deactivate, edit webhook URL |
| **Payments** | Monitor all payments, filter by status/chain/merchant, full transaction detail |
| **Payouts** | Track outgoing payouts, approval status, transaction hashes |
| **Refunds** | View refund requests, processing status, error tracking |
| **Batch Payouts** | Batch operation overview with per-item breakdown and progress |
| **Wallets** | Browse all HD-derived addresses across chains and merchants |
| **Balances** | Merchant available/pending balance per chain and token |
| **Currencies** | View and manage supported currencies and tokens per chain |
| **Chain Status** | Real-time blockchain sync state and block height per chain |
| **Admin Users** | RBAC management — create/delete admins, assign roles (`super_admin` / `admin` / `viewer`) |
| **Security** | JWT auth, login rate limiting, CORS whitelist, HSTS, CSP, SSRF protection |
| **i18n** | Error messages in 5 languages (en, zh, ja, ko, es) via `Accept-Language` header or `?lang=` query param |

### Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go + Gin + GORM + JWT + PostgreSQL + Redis (built into this repo) |
| Frontend | React 18 + TypeScript + Vite + Tailwind CSS v4 |
| Data | TanStack Query + TanStack Table + Recharts |
| State | Zustand + React Router v7 |

### Quick Start

```bash
# 1. Configure admin in config.yaml (see below)
# 2. Start OctopusWallet server (admin API included automatically)
make run-server

# 3. In another terminal, start the admin frontend
git clone https://github.com/gopvra/OctopusWallet-Admin.git
cd OctopusWallet-Admin
npm install && npm run dev
# Open http://localhost:5173
```

### Admin Configuration

```yaml
admin:
  jwt_secret: "your-secure-secret"     # JWT signing key (REQUIRED — server will not start without this)
  default_user: "admin"                 # Default admin username
  default_pass: "changeme"             # Default admin password (if empty, a random password is generated and logged)
  allowed_origins:                      # CORS origins for admin frontend
    - "http://localhost:5173"
    - "https://admin.yourdomain.com"
```

| Config | Environment Variable | Description |
|--------|---------------------|-------------|
| `admin.jwt_secret` | `OCTOPUS_ADMIN_JWT_SECRET` | JWT signing secret (required — server refuses to start if empty) |
| `admin.default_user` | `OCTOPUS_ADMIN_DEFAULT_USER` | Default admin username |
| `admin.default_pass` | `OCTOPUS_ADMIN_DEFAULT_PASS` | Default admin password (random generated if empty) |

### Default Credentials

On first startup, a default `super_admin` account is created:

```
Username: admin
Password: changeme
```

> **Important:** Change the default password immediately in production. If `default_pass` is left empty in the configuration, a secure random password will be generated and printed to the server log on first startup.

## License

Apache License 2.0
