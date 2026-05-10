package audit

import (
	"github.com/go-minstack/core"
	audit_repositories "github.com/ricardoalcantara/min-idp/internal/audit/repositories"
	"go.uber.org/fx"
)

func Register(app *core.App) {
	app.Provide(fx.Annotate(
		audit_repositories.NewAuditRepository,
		fx.As(new(AuditRepository)),
	))
	app.Provide(NewAuditService)
	app.Provide(NewAuditController)
}
