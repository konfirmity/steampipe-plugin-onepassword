---
title: "Steampipe Table: onepassword_audit_event - Query 1Password audit events using SQL"
description: "Allows users to query actions team members performed in a 1Password Business account through the 1Password Events API."
---

# Table: onepassword_audit_event - Query 1Password audit events using SQL

The 1Password Events API records actions team members take in a 1Password Business account, such as inviting or suspending users, changing vault access, and changing account security settings.

## Table Usage Guide

The `onepassword_audit_event` table lets security teams review administrative changes, for example multi-factor authentication or SSO being turned off.

**Important Notes**
- Requires `events_token` in the connection configuration: a bearer token from a 1Password Events Reporting integration with access to audit events. Set `events_url` for accounts outside 1password.com.
- Returns the last 24 hours unless `timestamp` is constrained with `>`, `>=`, `<` or `<=`.

## Examples

### Security settings turned off in the last 30 days

```sql+postgres
select
  timestamp,
  action,
  object_type,
  actor_email
from
  onepassword_audit_event
where
  timestamp >= now() - interval '30 days'
  and action in ('disblmfa', 'disblduo', 'disblsso')
order by
  timestamp desc;
```

```sql+sqlite
select
  timestamp,
  action,
  object_type,
  actor_email
from
  onepassword_audit_event
where
  timestamp >= datetime('now', '-30 days')
  and action in ('disblmfa', 'disblduo', 'disblsso')
order by
  timestamp desc;
```

### Users suspended in the last 24 hours

```sql+postgres
select
  timestamp,
  actor_email,
  object_details ->> 'email' as suspended_user
from
  onepassword_audit_event
where
  action = 'suspend'
  and object_type = 'user';
```

```sql+sqlite
select
  timestamp,
  actor_email,
  json_extract(object_details, '$.email') as suspended_user
from
  onepassword_audit_event
where
  action = 'suspend'
  and object_type = 'user';
```
