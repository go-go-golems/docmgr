//go:build !linux && !darwin && !freebsd

package operations

import (
	"context"
	"fmt"
)

func lock(_ context.Context, _ string) (func(), error) {
	return nil, fmt.Errorf("ticket operation locking requires Linux, macOS or FreeBSD")
}
