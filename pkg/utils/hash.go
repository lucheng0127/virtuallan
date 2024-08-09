package utils

import (
	"hash/fnv"
)

func IdxFromString(step int, str string) int {
	h := fnv.New32a()
	h.Write([]byte(str))

	return int(h.Sum32() % uint32(step))
}
