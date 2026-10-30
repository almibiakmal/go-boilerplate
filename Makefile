.PHONY: run-http run-event migrate-up migrate-down migrate-version migrate-create swagger test build

run-http:        ; go run ./src http serve
run-event:       ; go run ./src event serve
migrate-up:      ; go run ./src migrate up
migrate-down:    ; go run ./src migrate down
migrate-version: ; go run ./src migrate version
migrate-create:  ; go run ./src migrate create $(name)
swagger:         ; swag init -d src -g main.go -o src/docs
test:            ; go test -v ./...
build:           ; go build -o bin/app ./src
