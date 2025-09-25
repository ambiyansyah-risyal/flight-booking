package sqlxrepo

import "strings"

// isUniqueViolation detects unique constraint errors from PostgreSQL drivers.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}

// isForeignKeyViolation detects foreign key constraint violations.
func isForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "foreign key constraint") || strings.Contains(msg, "violates foreign key")
}
