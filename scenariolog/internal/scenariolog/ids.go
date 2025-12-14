package scenariolog

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"
)

// NewRunID creates a sortable, reasonably unique run id without external deps.
//
// Format: <utc-timestamp>-pid-<pid>-<randhex>
// Example: 2025-12-13T22:03:04.123456789Z-pid-12345-1a2b3c4d5e6f7788
func NewRunID(now time.Time) string {
	ts := now.UTC().Format(time.RFC3339Nano)

	var b [8]byte
	_, _ = rand.Read(b[:]) // best-effort; zero bytes still produce a valid id
	suffix := hex.EncodeToString(b[:])

	return fmt.Sprintf("%s-pid-%d-%s", ts, os.Getpid(), suffix)
}


