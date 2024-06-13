run:
    GREENLIGHT_DB_DSN="postgres://greenlight:password@localhost/greenlight?sslmode=disable" go run ./cmd/api

migrate-up:
    migrate -path=./migrations -database="postgres://greenlight:password@localhost/greenlight?sslmode=disable" up
