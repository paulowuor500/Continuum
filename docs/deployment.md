# Deployment guide

## Recommended platform
Render is the simplest first deployment target for this project because it supports both the API container and the web container.

## Required environment variables
Set these in your deployment platform:

- DATABASE_URL
- PORT=8080
- ALLOW_DEV_TIME_WARP=false
- LND_HOST
- LND_MACAROON_PATH
- LND_TLS_CERT_PATH
- NEXT_PUBLIC_API_URL

## Notes
- The API expects a PostgreSQL instance with the schema from [apps/api/init.sql](../apps/api/init.sql).
- The Lightning integration uses the configured LND gRPC host and credentials.
- The frontend should point its API URL at the deployed API service.

## Deploy steps
1. Push the repository changes to GitHub.
2. Create a new Render project and connect the repository.
3. Render will detect [render.yaml](../render.yaml).
4. Configure the environment variables above.
5. Deploy the services.
