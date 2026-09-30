---
title: "Steampipe Table: onepassword_user - Query 1Password users using SQL"
description: "Allows users to query the users of a 1Password account, through the 1Password Users API."
---

# Table: onepassword_user - Query 1Password users using SQL

The 1Password Users API (public preview) lists the people in a 1Password account with their email address, display name, state and creation time.

## Table Usage Guide

The `onepassword_user` table lets security and compliance teams reconcile 1Password accounts against their HR or identity records, for example to find suspended users or accounts that belong to people who have left.

**Important Notes**
- Requires `users_client_id`, `users_client_secret` and `users_account_id` in the connection configuration: an OAuth client created in 1Password with the list users scope, and the ID of the account. Set `users_api_url` to override `https://api.1password.com`.
- The Users API is not available on accounts that use automated provisioning (SCIM).

## Examples

### List all users

```sql+postgres
select
  id,
  email,
  display_name,
  state,
  create_time
from
  onepassword_user;
```

```sql+sqlite
select
  id,
  email,
  display_name,
  state,
  create_time
from
  onepassword_user;
```

### Suspended users

```sql+postgres
select
  email,
  display_name
from
  onepassword_user
where
  state = 'SUSPENDED';
```

```sql+sqlite
select
  email,
  display_name
from
  onepassword_user
where
  state = 'SUSPENDED';
```

### Users created in the last 30 days

```sql+postgres
select
  email,
  create_time
from
  onepassword_user
where
  create_time >= now() - interval '30 days';
```

```sql+sqlite
select
  email,
  create_time
from
  onepassword_user
where
  create_time >= datetime('now', '-30 days');
```
