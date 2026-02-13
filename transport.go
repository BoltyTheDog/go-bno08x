package bno08x

import "context"

// Transport defines the interface for communicating with the BNO08x
type Transport interface {
	// Send writes a buffer to the device
	Send(ctx context.Context, data []byte) error
	// Receive reads from the device into the provided buffer
	Receive(ctx context.Context, data []byte) (int, error)
	// Close cleans up the transport
	Close() error
}
