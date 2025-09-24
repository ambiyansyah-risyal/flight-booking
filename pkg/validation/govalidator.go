package validation

import (
    gv "github.com/asaskevich/govalidator"
)

// IsNonEmpty checks a string is not empty after trimming spaces.
func IsNonEmpty(s string) bool {
    return !gv.IsNull(s)
}

