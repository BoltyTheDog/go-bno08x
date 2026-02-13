package bno08x

import (
	"context"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

type I2CTransport struct {
	dev *i2c.Dev
}

func NewI2CTransport(busName string, address uint16) (*I2CTransport, error) {
	if _, err := host.Init(); err != nil {
		return nil, err
	}

	bus, err := i2creg.Open(busName)
	if err != nil {
		return nil, err
	}

	return &I2CTransport{
		dev: &i2c.Dev{Bus: bus, Addr: address},
	}, nil
}

func (t *I2CTransport) Send(ctx context.Context, data []byte) error {
	return t.dev.Tx(data, nil)
}

func (t *I2CTransport) Receive(ctx context.Context, data []byte) (int, error) {
	if err := t.dev.Tx(nil, data); err != nil {
		return 0, err
	}
	return len(data), nil
}

func (t *I2CTransport) Close() error {
	// i2c.Bus doesn't have a Close in some contexts, but let's assume it might
	return nil
}
