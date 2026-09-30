package onepassword

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableOnepasswordUser(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "onepassword_user",
		Description: "Users in the 1Password account, from the Users API (public preview).",
		List: &plugin.ListConfig{
			Hydrate: listUsers,
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_STRING, Description: "The unique identifier of the user.", Transform: transform.FromField("ID")},
			{Name: "email", Type: proto.ColumnType_STRING, Description: "The email address of the user.", Transform: transform.FromField("Email")},
			{Name: "display_name", Type: proto.ColumnType_STRING, Description: "The display name of the user.", Transform: transform.FromField("DisplayName")},
			{Name: "state", Type: proto.ColumnType_STRING, Description: "The state of the user: ACTIVE or SUSPENDED.", Transform: transform.FromField("State")},
			{Name: "create_time", Type: proto.ColumnType_TIMESTAMP, Description: "When the user was created.", Transform: transform.FromField("CreateTime")},
		},
	}
}

func listUsers(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := getUsersClient(d)
	if err != nil {
		plugin.Logger(ctx).Error("onepassword_user.listUsers", "connection_error", err)
		return nil, err
	}
	err = client.listUsers(ctx, func(user apiUser) bool {
		d.StreamListItem(ctx, user)
		return d.RowsRemaining(ctx) != 0
	})
	if err != nil {
		plugin.Logger(ctx).Error("onepassword_user.listUsers", "api_error", err)
		return nil, err
	}
	return nil, nil
}
