package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"
)

const TICKET_PREFIX = "OTT"

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func base62Encode(n int64) string {
	if n == 0 {
		return string(alphabet[0])
	}
	var sb strings.Builder
	for n > 0 {
		remainder := n % 62
		sb.WriteByte(alphabet[remainder])
		n = n / 62
	}
	runes := []rune(sb.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func randomPart(length int) (string, error) {
	var sb strings.Builder
	for range length {
		num, err := rand.Int(rand.Reader, big.NewInt(62))
		if err != nil {
			return "", err
		}
		sb.WriteByte(alphabet[num.Int64()])
	}
	return sb.String(), nil
}

func GenerateTicket(dbID int64, ts time.Time) (string, error) {
	timePart := base62Encode(ts.Unix())

	idPart := base62Encode(dbID)

	randomSuffix, err := randomPart(5)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s-%s-%s", TICKET_PREFIX, timePart, idPart, randomSuffix), nil
}
