package app

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimitedWriter limits the rate of data written to the underlying writer.
type RateLimitedWriter struct {
	writer    io.Writer
	rateLimit int64
	lastWrite time.Time
	mu        sync.Mutex
}

// NewRateLimitedWriter creates a new RateLimitedWriter with the specified rate limit.
func NewRateLimitedWriter(writer io.Writer, rateLimit string) *RateLimitedWriter {
	limit, err := parseRateLimit(rateLimit)
	if err != nil {
		fmt.Printf("Invalid rate limit: %v\n", err)
		return nil
	}

	return &RateLimitedWriter{
		writer:    writer,
		rateLimit: limit,
		lastWrite: time.Now(),
	}
}

func (rlw *RateLimitedWriter) Write(p []byte) (n int, err error) {
	rlw.mu.Lock()
	defer rlw.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rlw.lastWrite)
	expectedDelay := time.Duration(float64(len(p)) / float64(rlw.rateLimit) * float64(time.Second))

	if elapsed < expectedDelay {
		time.Sleep(expectedDelay - elapsed)
	}

	n, err = rlw.writer.Write(p)
	rlw.lastWrite = time.Now()
	return
}

func parseRateLimit(rateLimit string) (int64, error) {
	rateLimit = strings.TrimSpace(rateLimit)
	if rateLimit == "" {
		return 0, fmt.Errorf("empty rate limit")
	}

	suffix := strings.ToLower(rateLimit[len(rateLimit)-1:])
	value := rateLimit[:len(rateLimit)-1]

	var multiplier int64
	switch suffix {
	case "k":
		multiplier = 1024
	case "m":
		multiplier = 1024 * 1024
	default:
		return 0, fmt.Errorf("invalid rate limit suffix: %s", suffix)
	}

	rate, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}

	return rate * multiplier, nil
}
