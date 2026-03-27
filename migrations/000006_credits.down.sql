-- Drop post production company junction table
DROP TRIGGER post_production_company_timestamp_update ON post_production_company;
DROP TABLE post_production_company;

-- Drop post person junction table
DROP TRIGGER post_person_timestamp_update ON post_person;
DROP TABLE post_person;

-- Drop production company table
DROP TRIGGER production_company_timestamp_update ON production_company;
DROP TRIGGER production_company_search_vector_update ON production_company;
DROP FUNCTION update_production_company_search_vector();
DROP TABLE production_company;

-- Drop person table
DROP TRIGGER person_timestamp_update ON person;
DROP TRIGGER person_search_vector_update ON person;
DROP FUNCTION update_person_search_vector();
DROP TABLE person;

-- Remove new columns from post table
ALTER TABLE post
    DROP COLUMN release_year,
    DROP COLUMN country_of_origin,
    DROP COLUMN language;
    