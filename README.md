# OctopusWallet

Open-source multi-chain merchant payment gateway, similar to BitPay. Supports automatic payment receiving, transaction monitoring, and merchant payouts across multiple blockchains.

## Supported Chains

| Chain | Native Coin | Token Standard | Address Format |
|-------|-------------|----------------|----------------|
| Ethereum | ETH | ERC-20 | 0x... |
| BSC | BNB | BEP-20 | 0x... |
| Polygon | MATIC | ERC-20 | 0x... |
| Solana | SOL | SPL Token | Base58 |
| TRON | TRX | TRC-20 | Base58Check |
| Bitcoin | BTC | - | Bech32 (Segwit) |

## Architecture

```
                    ┌─────────────┐
  Merchant ───────> │  API Server │ ──── PostgreSQL
                    └─────────────┘
                    ┌─────────────┐
  Blockchains ───> │   Worker    │ ──── PostgreSQL
                    └─────────────┘
                          │
                    Webhook ──────> Merchant
```

- **API Server** (`cmd/server`): REST API for merchant registration, payment creation, payout requests
- **Worker** (`cmd/worker`): Background blockchain monitoring, payment detection, webhook delivery, payout processing

## Quick Start

### Prerequisites

- Go 1.22+
- PostgreSQL 16+
- Docker & Docker Compose (optional)

### 1. Start PostgreSQL

```bash
docker-compose up -d postgres
```

### 2. Configure

```bash
cp config/config.example.yaml config/config.yaml
# Edit config.yaml with your RPC endpoints and master seed
```

### 3. Run Database Migration

```bash
DATABASE_URL="postgres://octopus:octopus@localhost:5432/octopus_wallet?sslmode=disable" make migrate
```

### 4. Run

```bash
make run-server   # Start API server on :8080
make run-worker   # Start blockchain monitor + payout processor
```

### Docker

```bash
docker-compose up -d
```

## API Endpoints

### Public

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/v1/merchants/register` | Register new merchant |

### Authenticated (X-API-Key header)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/merchants/profile` | Get merchant profile |
| POST | `/api/v1/payments/create` | Create payment request |
| GET | `/api/v1/payments/:id` | Get payment status |
| POST | `/api/v1/payouts/create` | Create payout/withdrawal |
| GET | `/api/v1/payouts/:id` | Get payout status |
| GET | `/api/v1/wallets` | List merchant wallets |

### Example: Register Merchant

```bash
curl -X POST http://localhost:8080/api/v1/merchants/register \
  -H "Content-Type: application/json" \
  -d '{"name": "My Store", "email": "store@example.com", "webhook_url": "https://mystore.com/webhook"}'
```

### Example: Create Payment

```bash
curl -X POST http://localhost:8080/api/v1/payments/create \
  -H "Content-Type: application/json" \
  -H "X-API-Key: oct_your_api_key" \
  -d '{"chain": "ethereum", "amount": "50000000000000000", "token": ""}'
```

## Payment Flow

1. Merchant calls `POST /payments/create` with chain and amount
2. System derives a fresh HD wallet address for this merchant
3. Returns payment address to merchant (merchant shows to customer)
4. Worker monitors blockchain for incoming transactions
5. Payment status: `pending` -> `confirming` -> `completed`
6. Webhook notifications sent at each status change

## Webhook Events

| Event | Description |
|-------|-------------|
| `payment.confirming` | Transaction detected, waiting for confirmations |
| `payment.completed` | Required confirmations reached |
| `payment.expired` | Payment expired (30 min default) |
| `payout.completed` | Payout transaction confirmed |
| `payout.failed` | Payout failed |

Webhooks include HMAC-SHA256 signature in `X-Webhook-Signature` header.

## Configuration

Set via `config/config.yaml` or environment variables (prefix `OCTOPUS_`):

| Config | Env Var | Description |
|--------|---------|-------------|
| `wallet.master_seed` | `OCTOPUS_WALLET_MASTER_SEED` | BIP-39 mnemonic (24 words) |
| `wallet.encryption_key` | `OCTOPUS_WALLET_ENCRYPTION_KEY` | 32-byte hex AES key |
| `database.url` | `OCTOPUS_DATABASE_URL` | PostgreSQL connection string |
| `chains.<name>.rpc_url` | - | Chain RPC endpoint |

## Security

- HD wallet: single BIP-39 seed derives all addresses deterministically
- API keys: SHA-256 hashed, never stored in plaintext
- Webhook signatures: HMAC-SHA256 for payload verification
- Key encryption: AES-256-GCM for private key material at rest

## License

Apache License 2.0
