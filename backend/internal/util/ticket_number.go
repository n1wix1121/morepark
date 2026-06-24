package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

func GenerateTicketNumber() string {
	now := time.Now()
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		n = big.NewInt(time.Now().UnixNano() % 10000)
	}
	return fmt.Sprintf("MP-%s-%04d", now.Format("20060102"), n.Int64())
}
