---
title: "Steampipe Table: onepassword_item - Query 1Password Items using SQL"
description: "Allows users to query 1Password Items, specifically retrieving details about individual items stored in the 1Password vaults."
---

# Table: onepassword_item - Query 1Password Items using SQL

1Password is a password management service that stores sensitive information, including passwords, software licenses, and notes, in a secure and encrypted format. It allows users to create and manage multiple vaults, each containing different sets of items. These items can range from login credentials, secure notes, credit card information, to identities.

## Table Usage Guide

The `onepassword_item` table provides insights into 1Password Items within 1Password. As a security analyst, explore item-specific details through this table, including categories, vault IDs, and associated metadata. Utilize it to uncover information about items, such as their categories, the vaults they belong to, and the intricacies of their details.

**Important Notes**
- This build reads item metadata only. Passwords, fields, sections, files and other item contents are never requested, and there are no per-category item tables; use `category` instead.
- `favorite`, `version`, `trashed` and `last_edited_by` are only populated when connecting to a Connect server; they are null with a service account.

## Examples

### Basic info
Explore which items in your OnePassword vault have been recently updated or edited. This can help you keep track of changes and ensure your data remains secure.

```sql+postgres
select
  id,
  vault_id,
  title,
  created_at,
  updated_at,
  last_edited_by,
  tags
from
  onepassword_item;
```

```sql+sqlite
select
  id,
  vault_id,
  title,
  created_at,
  updated_at,
  last_edited_by,
  tags
from
  onepassword_item;
```

### List items that have been updated in the last 30 days
Explore which items have been modified in the past month to keep track of recent changes.

```sql+postgres
select
  id,
  vault_id,
  title,
  created_at,
  updated_at,
  last_edited_by,
  tags
from
  onepassword_item
where
  updated_at > now() - interval '30 day';
```

```sql+sqlite
select
  id,
  vault_id,
  title,
  created_at,
  updated_at,
  last_edited_by,
  tags
from
  onepassword_item
where
  updated_at > datetime('now', '-30 day');
```

### List items with production tag
Identify instances where items are tagged as 'production' to focus on operational elements.

```sql+postgres
select
  id,
  title,
  category,
  version,
  tags
from
  onepassword_item
where
  tags @> '["production"]';
```

```sql+sqlite
Error: The corresponding SQLite query is unavailable.
```

### Count items by category and state
Summarise what kinds of items exist, and how many are archived.

```sql+postgres
select
  category,
  state,
  count(*) as items
from
  onepassword_item
group by
  category,
  state
order by
  items desc;
```

```sql+sqlite
select
  category,
  state,
  count(*) as items
from
  onepassword_item
group by
  category,
  state
order by
  items desc;
```
