---
title: "Steampipe Table: onepassword_signin_attempt - Query 1Password sign-in attempts using SQL"
description: "Allows users to query sign-in attempts to a 1Password Business account through the 1Password Events API."
---

# Table: onepassword_signin_attempt - Query 1Password sign-in attempts using SQL

The 1Password Events API reports every successful and failed attempt to sign in to a 1Password Business account, with the user, the client and the approximate location.

## Table Usage Guide

The `onepassword_signin_attempt` table lets security teams review failed sign-ins, repeated MFA failures and blocked attempts.

**Important Notes**
- Requires `events_token` in the connection configuration: a bearer token from a 1Password Events Reporting integration with access to sign-in attempts. Set `events_url` for accounts outside 1password.com.
- Returns the last 24 hours unless `timestamp` is constrained with `>`, `>=`, `<` or `<=`.

## Examples

### Failed sign-ins in the last 7 days

```sql+postgres
select
  timestamp,
  category,
  type,
  target_user_email,
  client_ip_address,
  country
from
  onepassword_signin_attempt
where
  timestamp >= now() - interval '7 days'
  and category <> 'success'
order by
  timestamp desc;
```

```sql+sqlite
select
  timestamp,
  category,
  type,
  target_user_email,
  client_ip_address,
  country
from
  onepassword_signin_attempt
where
  timestamp >= datetime('now', '-7 days')
  and category <> 'success'
order by
  timestamp desc;
```

### Users with ten or more failed sign-ins in the last 24 hours

```sql+postgres
select
  target_user_email,
  count(*) as failures
from
  onepassword_signin_attempt
where
  category in ('credentials_failed', 'mfa_failed')
group by
  target_user_email
having
  count(*) >= 10;
```

```sql+sqlite
select
  target_user_email,
  count(*) as failures
from
  onepassword_signin_attempt
where
  category in ('credentials_failed', 'mfa_failed')
group by
  target_user_email
having
  count(*) >= 10;
```
