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

func IsRunOutOfIP(err error) bool {
	return strings.Contains(err.Error(), "run out of ip")
}

func IsTapNotExist(err error) bool {
	return strings.Contains(err.Error(), "file descriptor in bad state")
}
