//go:build !windows && !linux && !darwin

package ping

import (
	"context"
	"errors"
)

func (this *IcmpPing) ping_rootless(ctx context.Context) IPingResult {
	return this.errorResult(errors.New("not supported"))
}
