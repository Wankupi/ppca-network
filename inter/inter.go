package inter

import "context"

type InboundServer interface {
	Accept(ctx context.Context)
}
