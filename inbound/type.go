package inbound

import "context"

type InboundServer interface {
	Accept(ctx context.Context)
}
