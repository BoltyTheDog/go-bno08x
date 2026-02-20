package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/BoltyTheDog/go-bno08x"
)

func main() {
	// 1. Create I2C transport
	// Usually "/dev/i2c-1" on Raspberry Pi
	transport, err := bno08x.NewI2CTransport("1", 0x4A)
	if err != nil {
		log.Fatalf("failed to create transport: %v", err)
	}
	defer transport.Close()

	// 2. Create driver
	sensor := bno08x.NewBNO08X(transport)

	// 3. Reset and Initialize
	ctx := context.Background()
	sensor.SoftReset()
	time.Sleep(500 * time.Millisecond)

	// 4. Enable Rotation Vector (50ms interval)
	err = sensor.EnableFeature(bno08x.SensorReportRotationVector, 50000)
	if err != nil {
		log.Fatalf("failed to enable feature: %v", err)
	}

	fmt.Println("Starting BNO08x read loop...")

	for {
		err := sensor.Process(ctx)
		if err != nil {
			fmt.Printf("Error processing: %v\n", err)
			continue
		}

		if q, ok := sensor.GetQuaternion(); ok {
			p, r, y := q.ToEuler()
			heading := y
			if heading < 0 {
				heading += 360
			}
			fmt.Printf("\rHeading: %6.1f° | Pitch: %6.1f° | Roll: %6.1f°", heading, p, r)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
