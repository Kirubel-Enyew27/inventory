# Inventory API

A small Go/Gin inventory service backed by PostgreSQL and GORM.

## Configuration

- `DB_DSN`: PostgreSQL connection string. Defaults to `postgres://postgres:postgres@localhost:5432/inventory?sslmode=disable`.
- `PORT`: HTTP port. Defaults to `8080`.

## Endpoints

- `GET /health`: checks application and database readiness.
- `POST /products`: creates a product.
- `GET /products`: lists products with pagination metadata.
- `GET /products/:id`: returns one product.
- `PUT /products/:id`: replaces a product.
- `PATCH /products/:id/stock`: adjusts stock by a positive or negative `delta`.
- `DELETE /products/:id`: deletes a product.

Product listing supports:

- `category`: exact category filter.
- `low_stock`: boolean filter for products at or below the low-stock threshold.
- `sku`: exact SKU filter.
- `q`: search text matched against SKU, name, and description.
- `limit`: page size from 1 to 100. Defaults to 20.
- `offset`: zero-based result offset. Defaults to 0.

Product `sku` values are required and must be unique.

Example stock adjustment:

```sh
curl -X PATCH http://localhost:8080/products/1/stock \
  -H 'Content-Type: application/json' \
  -d '{"delta":-3}'
```

## Development

```sh
go test ./...
go run ./cmd/server
```
