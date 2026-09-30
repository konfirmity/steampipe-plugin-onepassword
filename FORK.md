# Konfirmity fork

This is konfirmity's build of the Steampipe 1Password plugin. It differs from upstream on purpose:
it reads **metadata only**. No code path returns item contents (passwords, fields, sections, files
or usernames), so a credential leak or a mistaken query cannot expose secrets.

## Divergence from upstream

- `GetItem` and the file calls are gone from the backend interface and both implementations.
- `onepassword_item` has no `Get` config and no `password`, `fields`, `sections`, `files` or
  `username` columns.
- The per-category tables (`onepassword_item_login`, `_password`, `_api_credential`,
  `_credit_card`, `_identity`, `_medical_record`, `_secure_note`, `_software_license`, `_file`) are
  removed; use `onepassword_item.category`.
- `metadata_only_test.go` fails the build if any column hydrates beyond its table's List call.
- `make postgres` builds the standalone Postgres FDW extension (FDW v2.2.6).

## Upstream branches

The konfirmity branch merges three feature branches that are meant for upstream turbot PRs.

| Branch | Adds | Upstream PR |
|---|---|---|
| `feat/service-account-auth` | service account authentication through the 1Password SDK | TBD |
| `feat/events-api-tables` | `onepassword_signin_attempt`, `onepassword_audit_event`, `onepassword_item_usage` | TBD |
| `feat/users-api-table` | `onepassword_user` | TBD |

Releases are tagged `v<upstream>-konfirmity.<n>`.
