package onepassword

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func Plugin(ctx context.Context) *plugin.Plugin {
	p := &plugin.Plugin{
		Name: "steampipe-plugin-onepassword",
		ConnectionConfigSchema: &plugin.ConnectionConfigSchema{
			NewInstance: ConfigInstance,
		},
		DefaultTransform: transform.FromCamel(),
		TableMap: map[string]*plugin.Table{
			"onepassword_audit_event":    tableOnepasswordAuditEvent(ctx),
			"onepassword_item":           tableOnepasswordItem(ctx),
			"onepassword_item_usage":     tableOnepasswordItemUsage(ctx),
			"onepassword_signin_attempt": tableOnepasswordSigninAttempt(ctx),
			"onepassword_user":           tableOnepasswordUser(ctx),
			"onepassword_vault":          tableOnepasswordVault(ctx),
		},
	}
	return p
}
