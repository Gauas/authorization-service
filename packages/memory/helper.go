package memory

import "fmt"
import "time"

const DAY_FMT = "20060102"

func refreshKey(token string) string {
	return fmt.Sprintf("auth:refresh:%s", token)
}

func deviceIndexKey(userID interface{}, deviceID string) string {
	return fmt.Sprintf("auth:device:%s:%s", userID, deviceID)
}

func globalBlacklistKey() string {
	return "auth:blacklists"
}

func blacklistBucketKey(day time.Time) string {
	return fmt.Sprintf("%s:%s", globalBlacklistKey(), day.UTC().Format(DAY_FMT))
}
