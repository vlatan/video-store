# Pickle to JSON Conversion

Because `pickle` is a Python specific format and we're rewriting this app in Go, Go can't actually read the pickled `BYTEA` database columns/values so we need to convert those to `JSONB`. But because Postgres can't do this we need a Python script for conversion.

So when finally migrating the production data you dump the production data with the schema, you comment out the `./migrations:/docker-entrypoint-initdb.d` volume and restore the data here along with its original schema. Do the manual labor described in the `convert.py` script and dump the resulting data only.

Then, uncomment the `./migrations:/docker-entrypoint-initdb.d` volume, run the first migration to create the tables, restore just the data, and run the rest of the migration steps. If everything is okay dump the data with the schema and restore it to a new database in production.
