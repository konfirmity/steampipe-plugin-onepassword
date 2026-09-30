---
title: "Steampipe Table: onepassword_item_usage - Query 1Password item usage using SQL"
description: "Allows users to query how items in shared 1Password vaults are used, through the 1Password Events API."
---

# Table: onepassword_item_usage - Query 1Password item usage using SQL

The 1Password Events API reports when items in shared vaults are filled, revealed, copied, shared or exported, and by whom.

## Table Usage Guide

The `onepassword_item_usage` table lets security teams see who is using which shared credentials, and spot exports or reveals.

**Important Notes**
- Requires `events_token` in the connection configuration: a bearer token from a 1Password Events Reporting integration with access to item usage. Set `events_url` for accounts outside 1password.com.
- Returns the last 24 hours unless `timestamp` is constrained with `>`, `>=`, `<` or `<=`.

## Examples

### Items revealed or exported in the last 7 days

```sql+postgres
select
  timestamp,
  action,
  item_uuid,
  vault_uuid,
  user_email
from
  onepassword_item_usage
where
  timestamp >= now() - interval '7 days'
  and action in ('reveal', 'export');
```

```sql+sqlite
select
  timestamp,
  action,
  item_uuid,
  vault_uuid,
  user_email
from
  onepassword_item_usage
where
  timestamp >= datetime('now', '-7 days')
  and action in ('reveal', 'export');
```

### Most used items in the last 24 hours

```sql+postgres
select
  item_uuid,
  count(*) as uses
from
  onepassword_item_usage
group by
  item_uuid
order by
  uses desc
limit 10;
```

```sql+sqlite
select
  item_uuid,
  count(*) as uses
from
  onepassword_item_usage
group by
  item_uuid
order by
  uses desc
limit 10;
```
