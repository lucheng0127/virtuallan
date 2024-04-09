package config

func ValidatePort(port int) bool {
	if port > 1024 && port < 65535 {
		return true
	}

	return false
}
