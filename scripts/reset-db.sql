-- Reset database PostgreSQL dengan aman
-- Menghapus semua tabel, index, sequence, dan constraint

-- Drop dan recreate schema public
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;

-- Grant permissions ke public schema (agar bisa diakses user biasa)
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO public;

-- Pastikan extension UUID tersedia
CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; 