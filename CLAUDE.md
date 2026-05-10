# min-idp — Engineering Rules

This file defines the coding standards, patterns, and decisions for this project.
It is updated every time a new pattern is established or adjusted.

---

## Git

- **Never add `Co-Authored-By` trailers to commit messages.**

---

## Module Structure

Each domain module follows this layout:

```
internal/<module>/
├── module.go               # Register(app *core.App) — wires all providers and invokes routes
├── <name>.controller.go    # Controller struct with handler methods
├── <name>.routes.go        # RegisterRoutes — maps HTTP paths to controller methods
├── <name>.service.go       # Service struct with business logic
├── dto/                    # Request/response data transfer objects
│   └── *.go
├── entities/               # GORM entity structs
│   └── <name>.entity.go
└── repositories/           # Data access layer wrapping repository.Repository[T]
    └── <name>.repository.go
```

Infrastructure modules (`config`, `storage`, `kvstore`) expose `Module() fx.Option` instead of `Register(app *core.App)` and are passed directly to `core.New(...)`.

## Handler Rules

- **Handlers are always methods on a Controller struct.** Never use inline functions (`func(c *gin.Context) { ... }`) in `main.go`, `Invoke()`, or route registration.
- `main.go` only bootstraps the app: `core.New(...)`, `domain.Register(app)`, `app.Invoke(...)`, `app.Run()`.
- Route files only map paths to controller methods — no logic.

## Dependency Injection

- All constructors follow the `NewX(deps...) *X` pattern for Uber FX.
- Modules register themselves via `Register(app *core.App)` — no global state.
- Prefer the **provide + invoke** split for non-HTTP side effects:
  ```go
  app.Provide(pkg.NewWorker)   // registers *Worker in the container
  app.Invoke(pkg.Run)          // FX injects *Worker → Run(w *Worker) calls w.Start()
  ```
  This keeps invoke targets single-responsibility, avoids inline closures, and makes the type available to other providers if needed.

## Error Responses

- Use `web.NewErrorDto(err)` for all error JSON responses.
- Use `web.NewMessageDto(msg)` for simple success messages.
- **Never use `gin.H{...}` for responses** — even single-field responses need a dedicated DTO.

## Logging

The framework injects `*slog.Logger` via FX. Always log through it — **never use `fmt.Println` or `log.Print`**.

| Level | When to use |
|-------|-------------|
| `log.Error()` | Real unexpected failures: DB down, infra errors, panics |
| `log.Debug()` | Expected application-level errors returned to the caller: not found, validation failure, wrong password |
| `log.Info()` | Significant events that happen infrequently: server started, bootstrap completed, key rotated |
| `log.Warn()` | Unexpected but recoverable: deprecated path used, fallback triggered |

## Layer Isolation

- **Services must not import `gin` or `net/http`.** They receive plain Go values and return domain values or errors.
- Controllers are the only layer that knows about HTTP.
- This keeps the service layer independently testable without an HTTP server.

## Entities and DTOs

- **Entities never cross the HTTP boundary.** Controllers always map entity → DTO before writing the response.
- This prevents leaking internal fields (e.g. `password_hash`) and decouples DB schema from the API contract.
- Mapping is done via constructor functions on the DTO (e.g. `NewUserDto(u *User) UserDto`).
- **All HTTP responses must use DTOs, never `gin.H{...}`.**

## Sentinel Errors and Domain Errors

### Sentinel Errors (Data Layer)

Shared sentinel errors live in `internal/db/errors.go`:

```go
var ErrEntityNotFound = errors.New("not found")
var ErrEntityConflict  = errors.New("conflict")
```

- Never define per-module duplicates (`ErrUserNotFound`, etc.).
- Repositories map `gorm.ErrRecordNotFound` → `db.ErrEntityNotFound` before returning.
- Controllers use `errors.Is()` to map sentinel errors to HTTP status codes.

### Domain Errors (Business Logic)

Domain errors are module-scoped unexported vars for expected application outcomes:

```go
// In internal/authn/authn.service.go
var errInvalidCredentials = errors.New("invalid credentials")
```

- **Domain error** (show to user, log Debug): business rule violation, validation failure.
- **Infrastructure error** (log Error, don't show): DB failure, system error.

## No Global State

- No `init()` functions. No package-level variables except sentinel errors, constants, and compiled templates.
- Everything is wired through FX constructors.

## Repository Rules

- Repositories only translate between Go types and the database. No business logic.
- Business logic lives exclusively in the service layer.

## Testing

### Layers and scope

| Layer | Type | Approach |
|-------|------|----------|
| Service | Unit — no DB, no HTTP | hand-written mocks implementing a repository interface |
| Full HTTP stack | E2E — real sqlite/postgres | `httptest` + real DB bootstrapped via `core.New()` |

### Unit tests (services)

- Each service file defines an **unexported repository interface** for every repo it depends on.
- Service struct fields use the interface type, not the concrete repo.
- Mocks are **hand-written** in `_test.go` in the **same package** as the service.
- Test function naming: `TestServiceName_Method_Scenario`.
- Test files live next to the code: `auth.service_test.go` alongside `auth.service.go`.
- Use `github.com/stretchr/testify` — `assert` for soft checks, `require` for hard stops.

### Coverage

- Target: **75% on the service layer**.
- Run: `make coverage`

## Naming

- Files: `<name>.<layer>.go` (e.g. `authn.service.go`, `user.entity.go`)
- Packages: **snake_case**, singular where possible
- DTOs: suffix `Dto` (e.g. `LoginDto`, `LoginResponseDto`)
- Sentinel errors: `ErrEntity` prefix for data-layer errors, plain `err` prefix for domain errors
