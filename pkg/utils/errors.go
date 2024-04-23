package utils

import "strings"

func IsIPConflictErr(err error) bool {
	return strings.Contains(err.Error(), "used by other endpoint")
}
