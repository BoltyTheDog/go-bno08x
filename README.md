# go-bno08x

A standalone Go library for the BNO08x IMU (9-DOF) sensor, focusing on high reliability and modularity. This library implements the CEVA SH-2 protocol over SHTP (Sensor Hub Transport Protocol).

## Features

- **Modular Transport**: Abstract communication layer (I2C support included via `periph.io`).
- **Precision Rotation**: Full support for Rotation Vector and Game Rotation Vector reports.
- **Euler Conversion**: Built-in quaternion to degrees (Pitch, Roll, Yaw) conversion.
- **Low-Level Control**: Direct access to SHTP framing and SH-2 report configuration.

## Installation

```bash
go get github.com/GammaDron/go-bno08x
```

## Architecture

The library is split into three distinct layers:

1.  **Transport**: Interface for the physical bus (I2C/SPI).
2.  **SHTP (Sensor Hub Transport Protocol)**: Handles packet framing, length headers, and channel-specific sequence numbers.
3.  **SH-2**: Implements the sensor reports and command protocols (Set Feature, Product ID, etc.).

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"github.com/GammaDron/go-bno08x"
)

func main() {
	// Create I2C transport (Rethink for your specific bus)
	transport, _ := bno08x.NewI2CTransport("1", 0x4A)
	sensor := bno08x.NewBNO08X(transport)

	// Enable Rotation Vector at 20Hz (50000us)
	sensor.EnableFeature(bno08x.SensorReportRotationVector, 50000)

	for {
		sensor.Process(context.Background())
		if q, ok := sensor.GetQuaternion(); ok {
			p, r, y := q.ToEuler()
			fmt.Printf("Heading: %.1f°\n", y)
		}
	}
}
```

## Implementation Details

- **SHTP**: Uses 4-byte headers. Length includes the header itself. Top bit of length is ignored.
- **Fixed-Point Scaling**: Uses Q-point math (e.g., Q14 for rotation vectors) to provide high-precision floating point results.
- **Linux Support**: Optimized for Raspberry Pi using `periph.io` for I2C communication.

## License

MIT
