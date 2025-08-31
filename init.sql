-- Initial database setup for fix-track-bot
-- This file is automatically executed when the PostgreSQL container starts

-- Create extensions if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create indexes for better performance (GORM will create the tables)
-- These will be applied after GORM auto-migration

-- Note: GORM will handle table creation through auto-migration
-- This file is mainly for any additional setup or extensions
