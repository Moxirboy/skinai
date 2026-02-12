-- Auth system migration
-- Ensure password column exists (may already be added manually)
ALTER TABLE users ADD COLUMN IF NOT EXISTS password VARCHAR(255) DEFAULT '';

-- Ensure additional columns exist
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_email_verified BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS "isPremium" BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS score INTEGER DEFAULT 0;
