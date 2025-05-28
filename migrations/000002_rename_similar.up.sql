DO $$ 
BEGIN
   IF EXISTS (SELECT 1 FROM information_schema.columns 
              WHERE table_name = 'post' 
              AND column_name = 'similar') THEN
       ALTER TABLE post RENAME COLUMN "similar" TO related;
   END IF;
END $$;