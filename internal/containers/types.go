package containers

import (
	"context"
)

type Container interface {
	Terminate(ctx context.Context)
}
