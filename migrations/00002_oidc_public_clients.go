package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pressly/goose/v3"
)

// Go returns programmatic migrations registered alongside embedded SQL.
func Go() []*goose.Migration {
	return []*goose.Migration{
		goose.NewGoMigration(2,
			&goose.GoFunc{RunDB: upOIDCPublicClients},
			&goose.GoFunc{RunDB: downOIDCPublicClients},
		),
	}
}

func upOIDCPublicClients(ctx context.Context, db *sql.DB) error {
	switch dbDriver(ctx, db) {
	case "sqlite":
		nullable, err := sqliteColumnNullable(ctx, db, "oidc_clients", "client_secret_hash")
		if err != nil {
			return err
		}
		if nullable {
			return nil
		}
		return sqliteRecreateOIDCClients(ctx, db, false)
	case "postgres":
		_, err := db.ExecContext(ctx,
			`ALTER TABLE oidc_clients ALTER COLUMN client_secret_hash DROP NOT NULL`)
		return err
	case "mysql":
		_, err := db.ExecContext(ctx,
			`ALTER TABLE oidc_clients MODIFY client_secret_hash TEXT NULL`)
		return err
	default:
		return fmt.Errorf("unsupported database driver for migration 2")
	}
}

func downOIDCPublicClients(ctx context.Context, db *sql.DB) error {
	switch dbDriver(ctx, db) {
	case "sqlite":
		return sqliteRecreateOIDCClients(ctx, db, true)
	case "postgres":
		if _, err := db.ExecContext(ctx,
			`UPDATE oidc_clients SET client_secret_hash = '' WHERE client_secret_hash IS NULL`); err != nil {
			return err
		}
		_, err := db.ExecContext(ctx,
			`ALTER TABLE oidc_clients ALTER COLUMN client_secret_hash SET NOT NULL`)
		return err
	case "mysql":
		if _, err := db.ExecContext(ctx,
			`UPDATE oidc_clients SET client_secret_hash = '' WHERE client_secret_hash IS NULL`); err != nil {
			return err
		}
		_, err := db.ExecContext(ctx,
			`ALTER TABLE oidc_clients MODIFY client_secret_hash TEXT NOT NULL`)
		return err
	default:
		return fmt.Errorf("unsupported database driver for migration 2 down")
	}
}

func dbDriver(ctx context.Context, db *sql.DB) string {
	var v string
	if err := db.QueryRowContext(ctx, `SELECT sqlite_version()`).Scan(&v); err == nil {
		return "sqlite"
	}
	if err := db.QueryRowContext(ctx, `SELECT version()`).Scan(&v); err == nil {
		if strings.Contains(strings.ToLower(v), "postgresql") {
			return "postgres"
		}
	}
	if err := db.QueryRowContext(ctx, `SELECT @@version`).Scan(&v); err == nil {
		return "mysql"
	}
	return ""
}

func sqliteColumnNullable(ctx context.Context, db *sql.DB, table, column string) (bool, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false, err
		}
		if name == column {
			return notnull == 0, nil
		}
	}
	return false, fmt.Errorf("column %s not found on %s", column, table)
}

func sqliteRecreateOIDCClients(ctx context.Context, db *sql.DB, secretNotNull bool) error {
	secretCol := "client_secret_hash TEXT"
	if secretNotNull {
		secretCol = "client_secret_hash TEXT NOT NULL"
	}

	_, err := db.ExecContext(ctx, fmt.Sprintf(`
		CREATE TABLE oidc_clients_new (
			id                     INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at             DATETIME,
			updated_at             DATETIME,
			deleted_at             DATETIME,
			sp_id                  INTEGER NOT NULL UNIQUE,
			client_id              TEXT NOT NULL UNIQUE,
			%s,
			redirect_uris          TEXT NOT NULL DEFAULT '[]',
			grant_types            TEXT NOT NULL DEFAULT '["authorization_code"]',
			response_types         TEXT NOT NULL DEFAULT '["code"]',
			scopes                 TEXT NOT NULL DEFAULT '["openid"]',
			token_endpoint_auth    TEXT NOT NULL DEFAULT 'client_secret_basic',
			pkce_required          INTEGER NOT NULL DEFAULT 0
		);
		INSERT INTO oidc_clients_new (
			id, created_at, updated_at, deleted_at,
			sp_id, client_id, client_secret_hash,
			redirect_uris, grant_types, response_types, scopes,
			token_endpoint_auth, pkce_required
		)
		SELECT
			id, created_at, updated_at, deleted_at,
			sp_id, client_id, client_secret_hash,
			redirect_uris, grant_types, response_types, scopes,
			token_endpoint_auth, pkce_required
		FROM oidc_clients;
		DROP TABLE oidc_clients;
		ALTER TABLE oidc_clients_new RENAME TO oidc_clients;
		CREATE INDEX IF NOT EXISTS idx_oidc_clients_deleted_at ON oidc_clients(deleted_at);
	`, secretCol))
	return err
}
