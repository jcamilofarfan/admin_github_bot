package utils

import "os"

// obtner el valor de una variable de entorno
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
