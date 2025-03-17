package utilis

import (
	"math/rand"
	"strings"
	"time"
)

// Helper functions for common utilities

// IsEmpty checks if a string is empty or contains only whitespace
func IsEmpty(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}

// Contains checks if a slice contains a specific element
func Contains[T comparable](slice []T, element T) bool {
	for _, item := range slice {
		if item == element {
			return true
		}
	}
	return false
}

// GenerateRandomString creates a random string of specified length
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// FormatDateTime formats time.Time to a standard string format
func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
