package line

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// Verify checks that the X-Line-Signature header value matches the
// Base64-encoded HMAC-SHA256 of the raw request body using the channel secret.
func Verify(channelSecret, signature string, body []byte) bool {
	mac := hmac.New(sha256.New, []byte(channelSecret))
	mac.Write(body)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
