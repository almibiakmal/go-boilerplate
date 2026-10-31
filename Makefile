.PHONY: run-http run-event migrate-up migrate-down migrate-version migrate-force migrate-create seed swagger test test-cover build

run-http:        ; go run ./src http serve
run-event:       ; go run ./src event servehttps://github.com/almibiakmal/go-boilerplate
migrate-up:      ; go run ./src migrate up
migrate-down:    ; go run ./src migrate down
migrate-version: ; go run ./src migrate version
migrate-force:   ; go run ./src migrate force $(v)
migrate-create:  ; go run ./src migrate create $(name)
seed:            ; go run ./src seed
swagger:         ; swag init -d src -g main.go -o src/docs
test:            ; go test -v ./...
test-cover:      ; go test -cover ./src/internal/handler ./src/internal/service ./src/internal/middleware ./src/internal/pkg/validator ./src/internal/infra/database ./src/internal/infra/logger
build:           ; go build -o bin/app ./src
