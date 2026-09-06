package utils

import "database/sql"

// ToNullInt64 is a helper function to convert
// an int64 to sql.NullInt64 on db UPDATE/INSERT
func ToNullInt64(i int64) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: i, Valid: true}
}

// ToNullInt16 is a helper function to convert
// an int16 to sql.NullInt16 on db UPDATE/INSERT
func ToNullInt16(i int16) sql.NullInt16 {
	if i == 0 {
		return sql.NullInt16{Valid: false}
	}
	return sql.NullInt16{Int16: i, Valid: true}
}

// ToNullString is a helper function to convert
// a string to sql.NullString on db UPDATE/INSERT
func ToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
