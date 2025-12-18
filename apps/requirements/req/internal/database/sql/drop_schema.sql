-- Drop all objects in specified schemas
DO $$ DECLARE
    schema_name TEXT;
    r RECORD;
    schemas TEXT[] := ARRAY['public'];
BEGIN
    FOREACH schema_name IN ARRAY schemas LOOP
        -- Drop tables first
        FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = schema_name) LOOP
            EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(schema_name) || '.' || quote_ident(r.tablename) || ' CASCADE';
        END LOOP;
        
        -- Then drop custom types (composites, domains, enums, ranges, multiranges), excluding system/array types
        FOR r IN (
            SELECT 
                t.typname,
                t.typtype,
                t.typisdefined
            FROM 
                pg_catalog.pg_type t
            JOIN 
                pg_catalog.pg_namespace n ON t.typnamespace = n.oid
            WHERE 
                n.nspname = schema_name
                AND t.typtype IN ('c', 'd', 'e', 'r', 'm')  -- c: composite, d: domain, e: enum, r: range, m: multirange
                AND t.typname NOT LIKE '\_%'  -- Exclude array types (which start with '_') They're dropped automatically with their base types.
                AND t.typisdefined     -- Only defined types
        ) LOOP
            EXECUTE 'DROP TYPE IF EXISTS ' || quote_ident(schema_name) || '.' || quote_ident(r.typname) || ' CASCADE';
        END LOOP;
    END LOOP;
END $$;
