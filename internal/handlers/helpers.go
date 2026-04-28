package handlers

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dtrabandt/payloadBuddy/internal/scenarios"
)

func getDurationParam(r *http.Request, param string, defaultValue time.Duration) time.Duration {
	val := r.URL.Query().Get(param)
	if val == "" {
		return defaultValue
	}
	if d, err := time.ParseDuration(val); err == nil {
		return d
	}
	if ms, err := strconv.Atoi(val); err == nil {
		return time.Duration(ms) * time.Millisecond
	}
	return defaultValue
}

func getIntParam(r *http.Request, param string, defaultValue int) int {
	val := r.URL.Query().Get(param)
	if val == "" {
		return defaultValue
	}
	if v, err := strconv.Atoi(val); err == nil {
		return v
	}
	return defaultValue
}

func getDelayStrategy(r *http.Request) scenarios.DelayStrategy {
	switch strings.ToLower(r.URL.Query().Get("strategy")) {
	case "fixed":
		return scenarios.FixedDelay
	case "random":
		return scenarios.RandomDelay
	case "progressive":
		return scenarios.ProgressiveDelay
	case "burst":
		return scenarios.BurstDelay
	default:
		return scenarios.FixedDelay
	}
}

func generateSysID() string {
	const chars = "abcdef0123456789"
	result := make([]byte, 32)
	for i := range result {
		idx, err := secureRandIntn(len(chars))
		if err != nil {
			result[i] = chars[i%len(chars)]
		} else {
			result[i] = chars[idx]
		}
	}
	return string(result)
}

func secureRandFloat32() (float32, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1<<24))
	if err != nil {
		return 0, err
	}
	return float32(n.Int64()) / float32(1<<24), nil
}

func secureRandIntn(n int) (int, error) {
	bigN, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0, err
	}
	return int(bigN.Int64()), nil
}

func secureRandInt63n(n int64) (int64, error) {
	bigN, err := rand.Int(rand.Reader, big.NewInt(n))
	if err != nil {
		return 0, err
	}
	return bigN.Int64(), nil
}
