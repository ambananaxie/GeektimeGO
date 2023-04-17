package main

import "context"

type MQ interface {
	Send(ctx context.Context, msg any) error
}
