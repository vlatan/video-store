package nulls

import "database/sql"

// Int64 converts int64 to sql.NullInt64
func Int64(i int64) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: i, Valid: true}
}

// Int16 converts int16 to sql.NullInt16
func Int16(i int16) sql.NullInt16 {
	if i == 0 {
		return sql.NullInt16{Valid: false}
	}
	return sql.NullInt16{Int16: i, Valid: true}
}

// String converts string to sql.NullString
func String(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
