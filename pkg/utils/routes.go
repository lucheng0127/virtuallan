package utils

import (
	"bufio"
	"bytes"
	"strings"
)

func ParseRoutesStream(data []byte) map[string]string {
	routes := make(map[string]string)
	reader := bytes.NewReader(data)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		rInfos := strings.Split(scanner.Text(), ">")
		if len(rInfos) == 2 {
			routes[rInfos[0]] = rInfos[1]
		}
	}

	return routes
}
