//go:build linux || darwin

package ping

import (
	"context"
)

func (this *IcmpPing) ping_rootless(ctx context.Context) IPingResult {
	return this.rawping("udp")
}
