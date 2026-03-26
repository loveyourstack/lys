# lyspgmon

Code for monitoring Postgres databases.

## lyspgmonddl

Views and associated stores for monitoring any Postgres database, including active queries, PG settings, table bloat and unused indexes.
Audit functions and stores.

Installation via the Install() func.

## LoveYourStack-specific checks

CheckDb() func, which reviews a Postgres database using LoveYourStack rules and conventions. These are:

1. If a table has an "updated_at" timestamp column, this will be updated via a trigger and not set manually. CheckDb() will add this trigger if it is missing.
1. If a table has a "last_user_update_by" column, the "t_audit_update" trigger will be added, which will store data changes in system.data_update.
1. Table shortnames should be set via a "shortname: " comment. CheckDb() checks that the shortname comments are unique.
1. If a table has an associated "_archived" table for archive (soft delete) functionality, CheckDb() checks that the base table columns and _archived table columns are consistent.
