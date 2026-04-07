# OctopusWallet

Open-source multi-chain merchant payment gateway. Self-hosted alternative to BitPay / CoinsPaid with enterprise features including auto-sweep, cold/hot wallet separation, withdrawal approval workflows, gas fee management, payment links, and audit logging.

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
  Merchant API ───> │  API Server  │ ──── PostgreSQL
  Dashboard UI ───> │  (+ Web UI)  │
                    └──────────────┘
                    ┌──────────────┐
  Blockchains ───> │   Worker     │ ──── PostgreSQL
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
# 1. Start PostgreSQL
docker-compose up -d postgres

# 2. Configure
cp config/config.example.yaml config/config.yaml

# 3. Run migrations
DATABASE_URL="postgres://octopus:octopus@localhost:5432/octopus_wallet?sslmode=disable" make migrate

# 4. Run backend
make run-server   # API on :8080
make run-worker   # Background services

# 5. Run frontend (development)
make web-install  # Install npm dependencies
make web-dev      # Vite dev server with HMR
```

### Docker (full stack)

```bash
docker-compose up -d   # Starts postgres + server + worker
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

Refund amount is validated against `amount_received` — cannot exceed the original payment.

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

All webhooks include `X-Webhook-Signature` (HMAC-SHA256) header for verification.

## Security Features

| Feature | Description |
|---------|-------------|
| **API Key Auth** | SHA-256 hashed keys, never stored in plaintext |
| **Request Signing** | Optional HMAC-SHA256 on requests (`X-Request-Signature`) |
| **Webhook Signing** | HMAC-SHA256 on all webhook payloads |
| **Idempotency** | `X-Idempotency-Key` header prevents duplicate requests |
| **IP Whitelist** | Per-merchant IP restriction, enforced in middleware |
| **Rate Limiting** | 100 req/s with 200 burst per connection |
| **Private Key Zeroing** | Key material wiped from memory after use |
| **HD Wallet** | BIP-39/32/44 deterministic address derivation |
| **Cold/Hot Separation** | Auto-transfer excess funds to cold storage |
| **Approval Workflow** | Configurable thresholds + daily/single-tx limits |
| **Atomic Processing** | SELECT FOR UPDATE SKIP LOCKED prevents double-processing |
| **Input Validation** | Amount (positive integer), address format (per-chain regex) |
| **Audit Log** | All mutating API calls recorded with IP + timestamp |
| **Refund Validation** | Refund amount cannot exceed payment received |
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
| Idempotency | ✓ | ✓ | ✓ |
| IP Whitelist | - | ✓ | ✓ |
| Request Signing | ✓ | ✓ | ✓ |
| Rate Limiting | - | - | ✓ |
| Supported Currencies API | ✓ | ✓ | ✓ |
| Balance/Ledger | ✓ | ✓ | ✓ |
| Pagination | ✓ | ✓ | ✓ |
| Web Dashboard | ✓ | ✓ | ✓ |
| Self-hosted | - | - | ✓ |
| Open Source | - | - | ✓ (Apache 2.0) |

## Configuration

Set via `config/config.yaml` or environment variables (prefix `OCTOPUS_`):

| Config | Env Var | Description |
|--------|---------|-------------|
| `wallet.master_seed` | `OCTOPUS_WALLET_MASTER_SEED` | BIP-39 mnemonic |
| `wallet.encryption_key` | `OCTOPUS_WALLET_ENCRYPTION_KEY` | 32-byte hex AES key |
| `database.url` | `OCTOPUS_DATABASE_URL` | PostgreSQL connection |
| `chains.<name>.rpc_url` | - | Chain RPC endpoint |
| `gas_station.enabled` | - | Enable gas fee management |
| `gas_station.chains.<name>.station_address` | - | Gas station address |

Frontend environment variable: `VITE_API_BASE` — override API base URL (default: `/api/v1`).

## Database Migrations

Migrations are in `migrations/` and run in order:

| File | Description |
|------|-------------|
| `001_init.sql` | Core tables: merchants, wallets, payments, payouts |
| `002_enterprise_features.sql` | Sweep, cold wallet, gas station, approval config |
| `003_invoices_refunds_ledger.sql` | Invoices, refunds, merchant balances, batch payouts |
| `004_payment_links_audit.sql` | Payment links, audit logs, payment tolerance fields |

## License

Apache License 2.0
