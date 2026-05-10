package audit

import (
	"github.com/go-minstack/core"
	audit_repositories "github.com/ricardoalcantara/min-idp/internal/audit/repositories"
)

func Register(app *core.App) {
	app.Provide(audit_repositories.NewAuditRepository)
	app.Provide(NewAuditService)
	app.Provide(NewAuditController)
}
