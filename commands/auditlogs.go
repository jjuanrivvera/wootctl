package commands

import "github.com/jjuanrivvera/wootctl/internal/api"

func init() {
	registerResource("", resourceSpec[api.Rec]{
		Use:     "audit-logs",
		Aliases: []string{"audit"},
		Short:   "Read the account audit log (Enterprise)",
		New:     func(c *api.Client) *api.Resource[api.Rec] { return c.AuditLogs() },
		Columns: []string{"id", "auditable_type", "action", "user_id", "created_at"},
		// Read-only resource: the audit trail can only ever be read.
		NoGet: true, NoCreate: true, NoUpdate: true, NoDelete: true,
	})
}
