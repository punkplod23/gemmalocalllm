-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Optionally, create a schema for ollama
CREATE SCHEMA IF NOT EXISTS ollama;

-- Create a table for storing vectors for use with Ollama
CREATE TABLE IF NOT EXISTS ollama.vects (
    id SERIAL PRIMARY KEY,
    "CompanyName" TEXT,
    "CompanyNumber" TEXT,
    "RegAddress.CareOf" TEXT,
    "RegAddress.POBox" TEXT,
    "RegAddress.AddressLine1" TEXT,
    "RegAddress.AddressLine2" TEXT,
    "RegAddress.PostTown" TEXT,
    "RegAddress.County" TEXT,
    "RegAddress.Country" TEXT,
    "RegAddress.PostCode" TEXT,
    "CompanyCategory" TEXT,
    "CompanyStatus" TEXT,
    "CountryOfOrigin" TEXT,
    "DissolutionDate" TEXT,
    "IncorporationDate" TEXT,
    "Accounts.AccountRefDay" TEXT,
    "Accounts.AccountRefMonth" TEXT,
    "Accounts.NextDueDate" TEXT,
    "Accounts.LastMadeUpDate" TEXT,
    "Accounts.AccountCategory" TEXT,
    "Returns.NextDueDate" TEXT,
    "Returns.LastMadeUpDate" TEXT,
    "Mortgages.NumMortCharges" TEXT,
    "Mortgages.NumMortOutstanding" TEXT,
    "Mortgages.NumMortPartSatisfied" TEXT,
    "Mortgages.NumMortSatisfied" TEXT,
    "SICCode.SicText_1" TEXT,
    "SICCode.SicText_2" TEXT,
    "SICCode.SicText_3" TEXT,
    "SICCode.SicText_4" TEXT,
    "LimitedPartnerships.NumGenPartners" TEXT,
    "LimitedPartnerships.NumLimPartners" TEXT,
    "URI" TEXT,
    "PreviousName_1.CONDATE" TEXT,
    "PreviousName_1.CompanyName" TEXT,
    "PreviousName_2.CONDATE" TEXT,
    "PreviousName_2.CompanyName" TEXT,
    "PreviousName_3.CONDATE" TEXT,
    "PreviousName_3.CompanyName" TEXT,
    "PreviousName_4.CONDATE" TEXT,
    "PreviousName_4.CompanyName" TEXT,
    "PreviousName_5.CONDATE" TEXT,
    "PreviousName_5.CompanyName" TEXT,
    "PreviousName_6.CONDATE" TEXT,
    "PreviousName_6.CompanyName" TEXT,
    "PreviousName_7.CONDATE" TEXT,
    "PreviousName_7.CompanyName" TEXT,
    "PreviousName_8.CONDATE" TEXT,
    "PreviousName_8.CompanyName" TEXT,
    "PreviousName_9.CONDATE" TEXT,
    "PreviousName_9.CompanyName" TEXT,
    "PreviousName_10.CONDATE" TEXT,
    "PreviousName_10.CompanyName" TEXT,
    "ConfStmtNextDueDate" TEXT,
    "ConfStmtLastMadeUpDate" TEXT,
    embedding VECTOR(1536) -- for vector search, adjust dimension as needed
);

-- Example: index for fast vector similarity search (optional, but recommended)
CREATE INDEX IF NOT EXISTS idx_vects_embedding ON ollama.vects USING ivfflat (embedding vector_cosine_ops) WITH (lists = 1000);

-- To load CSVs, use \copy from psql, not COPY or DO blocks in init.sql.
-- COPY cannot be used dynamically in a DO block.
-- Remove the DO block below to avoid errors.

-- DO $$ 
-- DECLARE
--     file record;
-- BEGIN
--     FOR file IN SELECT pg_ls_dir('/csv') AS filename
--     LOOP
--         IF file.filename ~ '\.csv$' THEN
--             EXECUTE format(
--                 $$COPY ollama.vects(text) FROM '/csv/%I' DELIMITER ',' CSV HEADER;$$,
--                 file.filename
--             );
--         END IF;
--     END LOOP;
-- END
-- $$;

-- Instead, after the database is running, connect with psql and run:
-- \copy ollama.vects(text) FROM '/csv/data.csv' DELIMITER ',' CSV HEADER;

-- When loading CSVs, the number and order of columns in the CSV must match the columns specified in the \copy or COPY command.
-- For example, if your CSV has more than one column, you must specify all columns or only those you want to import.

-- Example: If your CSV has columns: name, address, city, country, etc.
-- And you only want to import the first column as text:
-- \copy ollama.vects(text) FROM '/csv/data.csv' DELIMITER ',' CSV HEADER;

-- If you want to import multiple columns, adjust your table and command accordingly:
-- CREATE TABLE ollama.vects (
--     id SERIAL PRIMARY KEY,
--     name TEXT,
--     address TEXT,
--     city TEXT,
--     country TEXT,
--     embedding VECTOR(1536)
-- );
-- \copy ollama.vects(name, address, city, country) FROM '/csv/data.csv' DELIMITER ',' CSV HEADER;

-- To fix your error:
-- 1. Check your CSV file's header and number of columns.
-- 2. Match the columns in your \copy or COPY command to the CSV columns.
-- 3. Or, preprocess your CSV to only include the columns you want to import.

-- You can add more initialization SQL here as needed
-- You can add more initialization SQL here as needed
