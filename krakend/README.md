# KrakenD Gateway

This is the first integration phase for KrakenD. It only proxies two read-only routes and does not replace the existing APIs.

## Routes

- `GET http://localhost:8082/pos/v1/products` -> `http://pos-api:8080/v1/products`
- `GET http://localhost:8082/cms/v1/ads` -> `http://cms-api:8081/v1/ads`

Both routes forward the `Authorization` header to the backend services. JWT validation still happens in the existing Go APIs.

## Run

From the repository root:

```bash
docker compose up -d postgres pos-api cms-api krakend
```

Check the gateway container:

```bash
docker compose ps krakend
docker compose logs krakend
```

## Verify

Use an existing valid bearer token from the API login flow, then call:

```bash
curl -H "Authorization: Bearer <token>" http://localhost:8082/pos/v1/products
curl -H "Authorization: Bearer <token>" http://localhost:8082/cms/v1/ads
```

Expected behavior:

- The gateway listens on host port `8082`.
- Backend services remain directly available on `8080` and `8081`.
- Unauthorized or invalid-token responses are returned by the Go APIs, not by KrakenD.
