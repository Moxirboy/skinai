-- Rollback auth migration
-- These columns may have existed before, so be careful
ALTER TABLE users DROP COLUMN IF EXISTS score;
ALTER TABLE users DROP COLUMN IF EXISTS "isPremium";
ALTER TABLE users DROP COLUMN IF EXISTS is_email_verified;
ALTER TABLE users DROP COLUMN IF EXISTS is_active;
