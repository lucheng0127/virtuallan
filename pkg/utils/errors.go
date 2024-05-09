package utils

import "strings"

func IsIPConflictErr(err error) bool {
	return strings.Contains(err.Error(), "used by other endpoint")
}

func IsUnauthedErr(err error) bool {
	return strings.Contains(err.Error(), "unauthed client")
}

func IsUnsupportedPkt(err error) bool {
	return strings.Contains(err.Error(), "unsupported vl pkt")
}
