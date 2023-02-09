package utils

import "time"

const TimestampFormatLayout = "2006-01-02T15:04:05-07:00 MST"

// type Timestamp string

func StampTimeNow() string {
	// return time.Now().Format(TimestampFormatLayout)
	return time.Now().Format(time.RFC3339Nano)
}

func Now() time.Time {
	return time.Now()
}
