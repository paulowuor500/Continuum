# Production infrastructure

## Required services
- PostgreSQL for vault persistence
- Bitcoin Core for on-chain support
- LND for Lightning invoice generation

## Runtime environment
Set the following environment variables in the deployment platform:
- DATABASE_URL
- LND_HOST
- LND_MACAROON_PATH
- LND_TLS_CERT_PATH
- ALLOW_DEV_TIME_WARP=false

## Local production-like stack
Use docker-compose.prod.yml to run the database and Lightning dependencies locally.
