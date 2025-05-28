# Pickle to JSON Conversion

Because `pickle` is a Python specific format and we're rewriting this app in Go, Go can't actually read the pickled `BYTEA` database columns/values so we need to convert those to `JSONB`. But because Postgres can't do this we need a Python script for conversion.

