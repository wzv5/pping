// +build windows

package ping

import (
	"context"
)

func (this *IcmpPing) ping_rootless(ctx context.Context) IPingResult {
	return this.rawping("ip")
}
