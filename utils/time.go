package utils

import "time"

const TimestampFormatLayout = "2006-01-02T15:04:05-07:00 MST"

type Timestamp string

func StampTimeNow() Timestamp {
	t := time.Now()
	return Timestamp(t.Format(TimestampFormatLayout))
}

func PubnubToZulu(pubnubTime int64) time.Time {
	return time.Unix(pubnubTime/10000000, 0)
}
