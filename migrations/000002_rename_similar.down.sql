DO $$ 
BEGIN
   IF EXISTS (SELECT 1 FROM information_schema.columns 
              WHERE table_name = 'post' 
              AND column_name = 'related') THEN
       ALTER TABLE post RENAME COLUMN related TO "similar";
   END IF;
END $$;