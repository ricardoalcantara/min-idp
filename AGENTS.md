# go-minstack App Context

Use this file as the starting `AGENTS.md` in any new project built on go-minstack.

## Import Root

- The library import root is `github.com/go-minstack/go-minstack`
- Import only the packages the app actually uses
- Examples:
  - `github.com/go-minstack/go-minstack/core`
  - `github.com/go-minstack/go-minstack/gin`
  - `github.com/go-minstack/go-minstack/auth`
  - `github.com/go-minstack/go-minstack/sqlite`
  - `github.com/go-minstack/go-minstack/repository`
  - `github.com/go-minstack/go-minstack/web`
  - `github.com/go-minstack/go-minstack/cli`
  - `github.com/go-minstack/go-minstack/migration`

## Mental Model

- `core` is the bootstrap layer, not the whole framework
- `core.New(...)` is the main entry point for composing apps
- `core.New(...)` already includes:
  - core lifecycle wiring
  - `.env` loading through `godotenv.Load()`
  - logger setup via the `logger` module
- Do not add `core.Module()` manually when using `core.New(...)`
- In normal apps, pass infrastructure modules into `core.New(...)`
- Use `app.Provide(...)` for your own constructors
- Use `app.Invoke(...)` for startup functions, route registration, and migrations
- Keep package boundaries explicit: do not import MySQL/Postgres/SQLite packages unless the app actually uses them

## Module Selection

- Minimal app bootstrap only:
  - `core`
- HTTP API:
  - `core` + `gin`
- HTTP API with auth:
  - `core` + `gin` + `auth`
- HTTP API with database:
  - `core` + `gin` + one of `sqlite`, `mysql`, or `postgres`
- Shared HTTP response DTOs:
  - `web`
- CLI or one-shot script:
  - `core` + `cli`
- SQL migrations:
  - `core` + one DB module + `migration`
- Generic CRUD repository helpers:
  - `repository`

## Environment Variables

Use these env vars in apps built on go-minstack:

- HTTP:
  - `MINSTACK_HOST`
  - `MINSTACK_HTTP_PORT`
  - `MINSTACK_PORT`
  - `MINSTACK_CORS_ORIGIN`
- Database:
  - `MINSTACK_DB_URL`
- Auth:
  - `MINSTACK_JWT_SECRET`
  - `MINSTACK_JWT_PRIVATE_KEY`
  - `MINSTACK_JWT_PUBLIC_KEY`
  - `MINSTACK_JWKS_URL`
- Logging:
  - `MINSTACK_LOG_LEVEL`
  - `MINSTACK_LOG_FORMAT`

Notes:

- `sqlite` uses `:memory:` if `MINSTACK_DB_URL` is not set
- `mysql` and `postgres` require `MINSTACK_DB_URL`
- `MINSTACK_HTTP_PORT` takes precedence over `MINSTACK_PORT`
- `core.New(...)` loads `.env` automatically

## Recommended App Layout

```text
cmd/main.go
internal/
  users/
    entities/
    repositories/
    dto/
    user.service.go
    user.controller.go
    user.routes.go
    module.go
```

Each domain should usually expose a single `Register(app *core.App)` function.

## Main App Pattern

Keep `main.go` thin:

```go
package main

import (
    "github.com/go-minstack/go-minstack/core"
    mgin "github.com/go-minstack/go-minstack/gin"
    "github.com/go-minstack/go-minstack/sqlite"
)

func main() {
    app := core.New(
        mgin.Module(),
        sqlite.Module(),
    )

    users.Register(app)
    tasks.Register(app)

    app.Run()
}
```

## Domain Registration Pattern

```go
func Register(app *core.App) {
    app.Provide(NewUserRepository)
    app.Provide(NewUserService)
    app.Provide(NewUserController)
    app.Invoke(RegisterRoutes)
}
```

Use `app.Use(...)` when you need to attach raw `fx.Option` values.

## HTTP API Pattern

For web apps:

- pass `gin.Module()` to `core.New(...)`
- route handlers receive `*gin.Engine` through `app.Invoke(...)`
- `gin` handles:
  - engine creation
  - server lifecycle
  - request logging and recovery middleware
  - optional CORS from `MINSTACK_CORS_ORIGIN`

Example route registration:

```go
func RegisterRoutes(r *gin.Engine, c *UserController) {
    api := r.Group("/api/users")
    api.GET("", c.list)
    api.POST("", c.create)
}
```

## Auth Pattern

Use `auth` when the app needs JWT signing or validation.

- pass `auth.Module()` to `core.New(...)`
- inject `*auth.JwtService`
- use:
  - `auth.Authenticate(jwt)`
  - `auth.RequireRole("admin")`
  - `auth.ClaimsFromContext(ctx)`

Auth key loading priority:

1. `MINSTACK_JWT_PRIVATE_KEY`
2. `MINSTACK_JWKS_URL`
3. `MINSTACK_JWT_PUBLIC_KEY`
4. `MINSTACK_JWT_SECRET`

Simple auth API shape:

```go
app := core.New(
    mgin.Module(),
    auth.Module(),
)

app.Provide(NewUserController)
app.Invoke(RegisterRoutes)
app.Run()
```

## Database Pattern

Pick exactly one database module per app unless there is a strong reason not to:

- `sqlite.Module()`
- `mysql.Module()`
- `postgres.Module()`

Each provides `*gorm.DB` through FX. Use constructor injection:

```go
func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}
```

Driver notes:

- `sqlite` is simplest for local development and tests
- `sqlite` defaults to in-memory if no `MINSTACK_DB_URL` is set
- `mysql` uses a custom `mysql.UUID` type for `binary(16)` UUID storage
- `postgres` works well with `uuid.UUID`

## Repository Pattern

`repository` is a utility package, not an infrastructure module:

- it does not expose `Module()`
- pair it with one DB module
- register your repository with `app.Provide(...)`

Example with `uint` IDs:

```go
type UserRepository struct {
    *repository.Repository[User, uint]
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{repository.New[User, uint](db)}
}
```

Use repository helpers like:

- `Where(...)`
- `Order(...)`
- `Paginate(...)`
- `FindOne(...)`
- `Count(...)`
- `UpdatesByID(...)`

## Web DTO Pattern

Use `web` for shared HTTP response DTOs when they fit:

- `web.NewErrorDto(err)`
- `web.NewMessageDto("ok")`

`web` is intentionally small and dependency-free.

## Console App Pattern

Use `cli` for one-shot scripts, admin commands, data syncs, or backfills.

- pass `cli.Module()` to `core.New(...)`
- provide exactly one implementation of `cli.ConsoleApp`
- the process exits automatically when `Run(ctx)` returns
- return a non-nil error to exit with code 1

Example:

```go
type App struct {
    users *UserRepository
}

func NewApp(users *UserRepository) cli.ConsoleApp {
    return &App{users: users}
}

func (a *App) Run(ctx context.Context) error {
    return nil
}

func main() {
    app := core.New(
        cli.Module(),
        sqlite.Module(),
    )
    app.Provide(NewUserRepository)
    app.Provide(NewApp)
    app.Run()
}
```

## Migration Pattern

Use `migration` for SQL-file migrations:

- pass one DB module to `core.New(...)`
- pass `migration.Module(migrationsFS)` to `core.New(...)`
- invoke `migration.Run`
- embed SQL files with `embed.FS`

Example:

```go
app := core.New(
    sqlite.Module(),
    migration.Module(migrations.FS),
)

app.Invoke(migration.Run)
app.Run()
```

Migration notes:

- `migration.Module(...)` provides `*migration.Migrator`
- `migration.Run` applies all pending migrations
- dialect is inferred from the active GORM driver

## Logging Pattern

- `core.New(...)` already wires logging through the `logger` module
- inject `*slog.Logger` anywhere you need logging
- configure with:
  - `MINSTACK_LOG_LEVEL`
  - `MINSTACK_LOG_FORMAT`

Example:

```go
func NewUserService(log *slog.Logger, repo *UserRepository) *UserService {
    return &UserService{log: log, repo: repo}
}
```

## Testing Pattern

- use `app.Start(ctx)` and `app.Stop(ctx)` in tests
- prefer `sqlite` with `MINSTACK_DB_URL=:memory:` for app-level tests
- use `app.Invoke(...)` to capture `*gin.Engine` or run `AutoMigrate(...)` in tests

Example shape:

```go
app := core.New(mgin.Module(), sqlite.Module())
todos.Register(app)
app.Invoke(func(db *gorm.DB) error { return db.AutoMigrate(&Todo{}) })
require.NoError(t, app.Start(ctx))
defer app.Stop(ctx)
```

## Do Not Do This

- Do not put everything under `core`
- Do not hide registration in `init()`
- Do not rely on globals for app services
- Do not register routes from random constructors; prefer explicit `app.Invoke(RegisterRoutes)`
- Do not add DB imports to a package unless it actually uses them
- Do not treat `repository` like an infrastructure module
- Do not manually add `logger.Module()` or `core.Module()` in a normal `core.New(...)` app

## AI Guidance

When changing or generating code for a go-minstack app:

- prefer simple, direct Go code
- keep `main.go` small
- keep domain wiring in `internal/<domain>/module.go`
- use constructor injection everywhere
- keep comments minimal
- use explicit module composition instead of magic behavior
- when requirements are unclear, preserve the existing package boundaries

## Good References

- `task-api` for a fuller app using `auth`, `gin`, `sqlite`, `repository`, and `web`
- `todo-api` for a smaller CRUD example
- package READMEs in the go-minstack repo for focused module usage
