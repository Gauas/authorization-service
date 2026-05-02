package memory

import "fmt"

func refreshKey(token string) string {
	return fmt.Sprintf("auth:refresh:%s", token)
}

func deviceIndexKey(userID interface{}, deviceID string) string {
	return fmt.Sprintf("auth:device:%s:%s", userID, deviceID)
}

func blacklistKey(userID interface{}) string {
	return fmt.Sprintf("auth:blacklist:%s", userID)
}

func tokenSeqKey(userID interface{}) string {
	return fmt.Sprintf("auth:token:seq:%s", userID)
}
