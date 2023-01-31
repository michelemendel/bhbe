package utils

import "github.com/segmentio/ksuid"

// Unique IDs
// https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html

func GenerateUUID() string {
	return ksuid.New().String()
}
