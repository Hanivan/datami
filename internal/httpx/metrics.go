package httpx

import "time"

// SourceMetrics adapts a source-aware metrics callback to the httpx callback
// shape by injecting source. Returns nil when fn is nil (Client nil-guards it).
func SourceMetrics(source string, fn func(source, op string, dur time.Duration, err error)) func(op string, dur time.Duration, err error) {
	if fn == nil {
		return nil
	}
	return func(op string, dur time.Duration, err error) { fn(source, op, dur, err) }
}
