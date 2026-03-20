-- Rollback migration: Drop scan-related tables
-- Version: 002
-- Description: Remove EASM scanning tables

-- Drop tables in reverse order (children first, parents last)
-- This avoids foreign key constraint errors

DROP TABLE IF EXISTS whois_records CASCADE;
DROP TABLE IF EXISTS dns_records CASCADE;
DROP TABLE IF EXISTS subdomains CASCADE;
DROP TABLE IF EXISTS scan_jobs CASCADE;

/*
🎓 TEACHING NOTES - Rollback Migrations

=== WHY ROLLBACK? ===

1. Development:
   - Mistake in migration? Roll back and fix
   - Testing different schema designs
   - Clean slate for re-running migrations

2. Production:
   - Deployment issues (rollback release)
   - Performance problems with new schema
   - Critical bug in new code

=== DROP ORDER MATTERS ===

Wrong order:
```sql
DROP TABLE scan_jobs;  -- ❌ ERROR!
-- subdomains has foreign key to scan_jobs
```

Correct order:
```sql
DROP TABLE subdomains;  -- ✅ No foreign keys to this
DROP TABLE scan_jobs;    -- ✅ Now safe to drop
```

General rule: Reverse order of creation
- Create: parent → child
- Drop: child → parent

=== CASCADE KEYWORD ===

```sql
DROP TABLE scan_jobs CASCADE;
```

What CASCADE does:
- Drops dependent objects (indexes, constraints)
- PostgreSQL specific (safety feature)
- Without CASCADE: error if dependencies exist

Alternatives:
```sql
DROP TABLE scan_jobs;           -- Error if dependencies
DROP TABLE scan_jobs RESTRICT;  -- Same as above (explicit)
DROP TABLE scan_jobs CASCADE;   -- Drop with dependencies
```

=== IF EXISTS ===

```sql
DROP TABLE IF EXISTS scan_jobs;
```

Benefits:
- No error if table doesn't exist
- Idempotent (can run multiple times)
- Safe for scripts

Without IF EXISTS:
```sql
DROP TABLE scan_jobs;  -- ERROR if table doesn't exist
```

=== MIGRATION WORKFLOW ===

Development:
```bash
# Apply migration
psql -U postgres -d mini_asm -f 002_create_scan_tables.up.sql

# Test code with new schema
go test ./...

# Issue found? Rollback:
psql -U postgres -d mini_asm -f 002_create_scan_tables.down.sql

# Fix migration
vim 002_create_scan_tables.up.sql

# Re-apply
psql -U postgres -d mini_asm -f 002_create_scan_tables.up.sql
```

Production (with tool like golang-migrate):
```bash
# Apply
migrate -path migrations -database postgres://... up

# Rollback if needed
migrate -path migrations -database postgres://... down 1
```

=== DATA LOSS WARNING ===

⚠️  DANGER: Rolling back drops tables and ALL DATA!

Before rollback in production:
1. Backup database
2. Export important data
3. Coordinate with team
4. Have forward fix ready

Better approach:
- Additive migrations (add, don't change)
- Feature flags (disable in code, not schema)
- Keep old columns temporarily

Example - Safe schema change:
```sql
-- DON'T DO THIS:
ALTER TABLE assets DROP COLUMN old_field;  -- ❌ Data loss

-- DO THIS:
ALTER TABLE assets ADD COLUMN new_field TEXT;  -- ✅ Additive
-- Update code to use new_field
-- After verified working, drop old_field in later migration
```

=== TESTING ROLLBACKS ===

Always test down migrations:

```bash
# Apply
./migrate up

# Verify schema
psql -c "\dt"  # Show tables

# Rollback
./migrate down

# Verify clean slate
psql -c "\dt"  # Should be empty (or back to previous state)

# Re-apply
./migrate up

# Verify again
psql -c "\dt"
```

If up/down cycle works cleanly → good migration!

=== COMMON MISTAKES ===

1. Wrong drop order:
```sql
DROP TABLE scan_jobs;   -- ❌ Has foreign keys pointing to it
DROP TABLE subdomains;  -- ✅ Should be first
```

2. Missing CASCADE:
```sql
DROP TABLE assets;  -- ❌ Has many dependencies
DROP TABLE assets CASCADE;  -- ✅ Drops dependencies too
```

3. Not testing rollback:
- Write up migration ✅
- Write down migration ❌ (forgot)
- Production issue → no clean way to rollback

4. Forgetting IF EXISTS:
```sql
DROP TABLE temp_table;  -- ❌ Error if table never existed
-- This breaks automation scripts
```

Always write BOTH up and down migrations!
Never commit incomplete migrations!
*/
