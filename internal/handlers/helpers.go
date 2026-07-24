package handlers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dtrabandt/payloadBuddy/internal/scenarios"
)

// queryParser reads typed query parameters and records the first parse failure,
// so a handler can validate every parameter and then reject the request once.
// Unparseable values are never silently replaced by defaults — that hides client bugs.
type queryParser struct {
	r   *http.Request
	err error
}

func newQueryParser(r *http.Request) *queryParser {
	return &queryParser{r: r}
}

// Err returns the first parse failure encountered, or nil when all values were valid.
func (q *queryParser) Err() error {
	return q.err
}

func (q *queryParser) fail(param, value, reason string) {
	if q.err == nil {
		q.err = fmt.Errorf("invalid %s parameter %q: %s", param, value, reason)
	}
}

// Int returns the named parameter as an int, or defaultValue when it is absent.
func (q *queryParser) Int(param string, defaultValue int) int {
	val := q.r.URL.Query().Get(param)
	if val == "" {
		return defaultValue
	}
	v, err := strconv.Atoi(val)
	if err != nil {
		q.fail(param, val, "must be an integer")
		return defaultValue
	}
	return v
}

// Duration returns the named parameter as a duration, accepting either Go duration
// syntax ("100ms") or bare milliseconds ("100"). Negative durations are rejected.
func (q *queryParser) Duration(param string, defaultValue time.Duration) time.Duration {
	val := q.r.URL.Query().Get(param)
	if val == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		ms, msErr := strconv.Atoi(val)
		if msErr != nil {
			q.fail(param, val, `must be a duration such as "100ms" or a number of milliseconds`)
			return defaultValue
		}
		d = time.Duration(ms) * time.Millisecond
	}
	if d < 0 {
		q.fail(param, val, "must not be negative")
		return defaultValue
	}
	return d
}

// Bool returns the named parameter as a bool, or defaultValue when it is absent.
func (q *queryParser) Bool(param string, defaultValue bool) bool {
	val := q.r.URL.Query().Get(param)
	if val == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		q.fail(param, val, "must be true or false")
		return defaultValue
	}
	return b
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
