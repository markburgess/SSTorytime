-- Initialize SSTorytime database
-- This runs automatically when postgres container starts

-- Create unaccent extension for text search
CREATE EXTENSION IF NOT EXISTS unaccent;

-- Grant all privileges
GRANT ALL PRIVILEGES ON DATABASE sstoryline TO sstoryline;
