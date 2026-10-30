# Majoo Core Service Boilerplate (`maj-core-srvc`)

Boilerplate microservice **Golang** berstandar *production* untuk kebutuhan **RESTful API** (dan siap diperluas ke **Event-Driven**), dibangun dengan fokus pada performa tinggi, kemudahan pemeliharaan (*maintainability*), dan batas antar-layer yang tegas.

---

## Tech Stack Utama

- **Language**: Go `1.26+`
- **HTTP Framework**: [Fiber v2](https://github.com/gofiber/fiber) (`fasthttp` engine)
- **CLI & Command Runner**: [Cobra](https://github.com/spf13/cobra) (`http serve`, `event serve`, `migrate`)
- **Database ORM & Pooling**: [GORM](https://gorm.io/) dengan dynamic driver (`postgres`, `mysql`, `sqlite`) di atas `database/sql` connection pool
- **Database Migration**: [golang-migrate](https://github.com/golang-migrate/migrate)
- **Validation**: [go-playground/validator v10](https://github.com/go-playground/validator)
- **Authentication**: [golang-jwt v5](https://github.com/golang-jwt/jwt) (Bearer Token + RBAC) & Internal API Key (`X-API-Key`)
- **API Documentation**: [Swaggo / OpenAPI 2.0](https://github.com/swaggo/swag) (`GET /swagger/index.html`)
- **Logging**: Structured Context-Aware Logger berbasis `log/slog` (otomatis menginjeksi `request_id` & `user_id` di seluruh layer + GORM Slow Query Logger)

---

## Struktur Direktori & Arsitektur (3-Tier Decoupled)

```text
maj-core-srvc/
├── migrations/                              # File SQL migrasi (.up.sql & .down.sql)
├── src/
│   ├── main.go                              # Entrypoint aplikasi & anotasi global Swagger
│   ├── cmd/                                 # Definisi CLI Cobra
│   │   ├── root.go                          # Inisialisasi Config, Logger, & Signal Handling
│   │   ├── http.go                          # Command: `app http serve`
│   │   ├── event.go                         # Command: `app event serve`
│   │   └── migrate.go                       # Command: `app migrate [up|down|version|force|create]`
│   ├── docs/                                # Generated OpenAPI/Swagger files (swag init)
│   └── internal/
│       ├── bootstrap/
│       │   └── app.go                       # Dependency Injection wiring & Graceful Shutdown
│       ├── config/
│       │   └── config.go                    # Konfigurasi berbasis Environment Variables (.env)
│       ├── infra/
│       │   ├── database/
│       │   │   ├── database.go              # Dynamic GORM connection & sql.DB Connection Pool
│       │   │   └── gorm_logger.go           # Custom GORM logger terintegrasi slog + Slow Query detector
│       │   └── logger/
│       │       └── logger.go                # Context-aware slog (JSON/Text, Stdout/File)
│       ├── middleware/
│       │   ├── middleware.go                # Global middleware setup (RequestID, Recover, CORS)
│       │   ├── logger.go                    # HTTP Access Logger & RequestID context propagation
│       │   ├── error_handler.go             # Centralized Fiber ErrorHandler
│       │   ├── auth.go                      # JWT Bearer Auth & Role-Based Access Control (RBAC)
│       │   ├── api_key.go                   # Service-to-Service Auth (X-API-Key)
│       │   └── rate_limiter.go              # IP-based Rate Limiter
│       ├── router/
│       │   └── http.go                      # Registrasi seluruh endpoint HTTP
│       ├── handler/                         # Layer 1: HTTP Delivery (parse request, validate, return JSON)
│       ├── service/                         # Layer 2: Business Logic (transport-agnostic)
│       ├── repository/                      # Layer 3: Data Access (GORM queries)
│       ├── model/                           # Entity Database murni (GORM structs, tanpa tag JSON)
│       ├── dto/                             # Kontrak Data (terpisah tegas per tanggung jawab)
│       │   ├── request/                     # Struct penerima HTTP Request (JSON body / Query params)
│       │   ├── service/                     # Struct input/filter untuk Service Layer
│       │   └── response/                    # Struct output JSON Response + Mapper dari Model
│       └── pkg/
│           ├── response/                    # Standar JSON Envelope, AppError, & Mapping Response Code
│           └── validator/                   # Wrapper validator struct -> FieldError standar
├── .env.example
├── Dockerfile                               # Multi-stage Docker build (Alpine non-root)
├── docker-compose.yml                       # Local development stack (PostgreSQL + App)
└── Makefile                                 # Shortcut operasional & development
```

### Alur Data Antar-Layer

Untuk mencegah *leaky abstraction* (skema tabel database bocor ke JSON API atau layer `service` bergantung pada HTTP), alur data dipisah secara ketat:

```mermaid
flowchart LR
    Client["HTTP Client"] -->|JSON / Query| ReqDTO["dto/request"]
    ReqDTO -->|.ToServiceInput()| SvcDTO["dto/service"]
    SvcDTO --> Service["Service Layer"]
    Service <-->|model.Entity| Repo["Repository Layer (GORM)"]
    Service -->|ToResponse()| RespDTO["dto/response"]
    RespDTO -->|response.Envelope| Client
```

---

## Quick Start

### 1. Persiapan Environment
Salin file `.env.example` menjadi `.env`:
```bash
cp .env.example .env
```

> **Dynamic Database Driver (`DB_DRIVER`)**:
> Anda dapat mengubah `DB_DRIVER` di `.env` ke `postgres`, `mysql`, atau `sqlite`. Jika `DB_URL` dikosongkan, aplikasi akan merakit connection string secara otomatis dari variabel `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, dan `DB_NAME`.

### 2. Menjalankan Database (Docker Compose)
```bash
docker compose up -d postgres
```

### 3. Menjalankan Migrasi Database
```bash
# Menjalankan semua migrasi pending
make migrate-up

# Mengecek versi migrasi saat ini
make migrate-version

# Membuat file migrasi baru (.up.sql & .down.sql)
make migrate-create name=add_orders_table

# Rollback 1 langkah migrasi
make migrate-down
```

### 4. Menjalankan Server
```bash
# Menjalankan REST API Server
make run-http

# Menjalankan Event Worker
make run-event
```

Setelah server berjalan:
- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **Liveness Probe**: `GET http://localhost:8080/healthz`
- **Readiness & DB Pool Stats**: `GET http://localhost:8080/readyz`

---

## Standar Format Response HTTP & Mapping Code

Seluruh endpoint mengembalikan struktur JSON `response.Envelope` yang seragam melalui package `src/internal/pkg/response`:

### Response Sukses (`200 OK` / `201 Created`)
```json
{
  "success": true,
  "code": "SUCCESS",
  "message": "Product retrieved successfully",
  "data": {
    "id": 1,
    "sku": "SKU-001",
    "name": "Kopi Susu Aren",
    "price": 18000,
    "created_at": "2026-10-31T00:00:00Z",
    "updated_at": "2026-10-31T00:00:00Z"
  }
}
```

### Response List dengan Pagination (`200 OK`)
```json
{
  "success": true,
  "code": "SUCCESS",
  "message": "Products retrieved successfully",
  "data": [],
  "meta": {
    "page": 1,
    "per_page": 10,
    "total_items": 45,
    "total_pages": 5
  }
}
```

### Response Error / Validasi Gagal (`4xx` / `5xx`)
```json
{
  "success": false,
  "code": "VALIDATION_ERROR",
  "message": "Validation failed",
  "error": {
    "code": "VALIDATION_ERROR",
    "details": [
      {
        "field": "sku",
        "message": "sku is required"
      }
    ]
  }
}
```

### Daftar Mapping `response.Code` Bawaan

| Konstanta `response.Code` | Kode String | HTTP Status |
| :--- | :--- | :---: |
| `CodeSuccess` | `SUCCESS` | `200 OK` |
| `CodeCreated` | `CREATED` | `201 Created` |
| `CodeBadRequest` | `BAD_REQUEST` | `400 Bad Request` |
| `CodeUnauthorized` | `UNAUTHORIZED` | `401 Unauthorized` |
| `CodeForbidden` | `FORBIDDEN` | `403 Forbidden` |
| `CodeNotFound` | `NOT_FOUND` | `404 Not Found` |
| `CodeConflict` | `CONFLICT` | `409 Conflict` |
| `CodeDuplicateEntry` | `DUPLICATE_ENTRY` | `409 Conflict` |
| `CodeValidationError` | `VALIDATION_ERROR` | `422 Unprocessable Entity` |
| `CodeTooManyRequests` | `TOO_MANY_REQUESTS` | `429 Too Many Requests` |
| `CodeInternalError` | `INTERNAL_SERVER_ERROR` | `500 Internal Server Error` |
| `CodeDatabaseError` | `DATABASE_ERROR` | `500 Internal Server Error` |
| `CodeServiceUnavailable` | `SERVICE_UNAVAILABLE` | `503 Service Unavailable` |

> Di dalam `service` atau `handler`, cukup kembalikan `response.NewError(response.CodeNotFound, "Pesan custom")` atau `response.WrapError(response.CodeDatabaseError, "Pesan", err)`. Middleware `ErrorHandler` akan otomatis menerjemahkannya ke HTTP status code dan format JSON di atas.

---

## Panduan Menambahkan Modul / Endpoint Baru

1. **Buat Migrasi & Model**:
   - Buat file migrasi dengan `make migrate-create name=create_xxx_table`.
   - Buat entity GORM di `src/internal/model/<module>.go`.
2. **Buat Kontrak DTO**:
   - Request HTTP di `src/internal/dto/request/<module>_request_dto.go` (tambahkan tag `validate:"..."`).
   - Input Service di `src/internal/dto/service/<module>_service_dto.go`.
   - Response JSON di `src/internal/dto/response/<module>_response_dto.go`.
3. **Buat Repository, Service, dan Handler**:
   - Implementasikan query DB di `src/internal/repository/`.
   - Implementasikan logika bisnis di `src/internal/service/`.
   - Implementasikan handler Fiber + anotasi Swagger di `src/internal/handler/`.
4. **Daftarkan ke Router & Bootstrap**:
   - Hubungkan dependency di `src/internal/bootstrap/app.go` dan daftarkan endpoint di `src/internal/router/http.go`.
   - Jalankan `make swagger` untuk memperbarui dokumentasi OpenAPI.

---

## Testing & Build

```bash
# Menjalankan seluruh unit test
make test

# Re-generate dokumentasi Swagger
make swagger

# Build binary ke ./bin/app
make build
```

