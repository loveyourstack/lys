# lyspgmon

Code for monitoring Postgres databases.

## lyspgmonddl

Views and associated stores for monitoring any Postgres database, including active queries, PG settings, table bloat and unused indexes.

Installation via the Install() func.

## Lys-specific DDL checks

CheckDDL() func, which reviews a Postgres database using lys-specific rules. These are:

1. If a table has an "updated_at" timestamp column, this will be updated via a trigger and not set manually. CheckDDL() will add this trigger if it is missing.
1. If a table has a "t_audit_update" trigger, CheckDDL() checks that it has a "last_user_update_by" column so that the audit function knows which user made the change.
1. Table shortnames should be set via a "shortname: " comment. CheckDDL() checks that the shortname comments are unique.
1. If a table has an associated "_archived" table for soft delete functionality, CheckDDL() checks that the base table columns and _archived table columns are consistent.
