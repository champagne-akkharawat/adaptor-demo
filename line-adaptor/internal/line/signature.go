package line

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

func Verify(channelSecret string, body []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(channelSecret))
	mac.Write(body)
	computed := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(computed), []byte(signature))
}
