package storage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ricardoalcantara/min-idp/internal/config"
	"go.uber.org/fx"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
)

type Params struct {
	fx.In

	Config *config.Config
	LC     fx.Lifecycle
	Log    *slog.Logger
}

func NewDB(p Params) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch p.Config.DBDriver {
	case "sqlite", "sqlite3":
		dialector = sqlite.Open(p.Config.DBDSN)
	case "mysql":
		dialector = mysql.Open(p.Config.DBDSN)
	case "postgres", "postgresql":
		dialector = postgres.Open(p.Config.DBDSN)
	default:
		return nil, fmt.Errorf("storage: unsupported db driver %q", p.Config.DBDriver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("storage: open db: %w", err)
	}

	if p.Config.DBDriver == "sqlite" || p.Config.DBDriver == "sqlite3" {
		if err := db.Exec("PRAGMA journal_mode=WAL").Error; err != nil {
			return nil, fmt.Errorf("storage: sqlite pragma journal_mode: %w", err)
		}
		if err := db.Exec("PRAGMA foreign_keys=ON").Error; err != nil {
			return nil, fmt.Errorf("storage: sqlite pragma foreign_keys: %w", err)
		}
	}

	p.LC.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			p.Log.Info("closing database")
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDB.Close()
		},
	})

	return db, nil
}
