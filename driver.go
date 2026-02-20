package bno08x

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type BNO08X struct {
	transport Transport
	mu        sync.Mutex

	seqNumbers [6]uint8
	readings   map[uint8]interface{}
	accuracies map[uint8]int
	Debug      bool
}

func NewBNO08X(transport Transport) *BNO08X {
	return &BNO08X{
		transport:  transport,
		readings:   make(map[uint8]interface{}),
		accuracies: make(map[uint8]int),
	}
}

func (b *BNO08X) SoftReset() error {
	// Send reset command on EXE channel
	data := []uint8{0x01}
	return b.sendPacket(ChanExecutable, data)
}

func (b *BNO08X) EnableFeature(feature uint8, intervalMicroseconds uint32) error {
	data := make([]byte, 21)
	data[0] = ReportSetFeatureCommand
	data[1] = feature
	data[2] = 0 // Flags
	data[3] = 0 // Change sensitivity LSB
	data[4] = 0 // Change sensitivity MSB
	// Interval
	data[5] = uint8(intervalMicroseconds & 0xFF)
	data[6] = uint8((intervalMicroseconds >> 8) & 0xFF)
	data[7] = uint8((intervalMicroseconds >> 16) & 0xFF)
	data[8] = uint8((intervalMicroseconds >> 24) & 0xFF)
	// Batch interval (0)
	// Sensor specific config (0)

	return b.sendPacket(ChanControl, data)
}

func (b *BNO08X) Process(ctx context.Context) error {
	// 1. Read header (4 bytes) to get packet length
	headerBuf := make([]byte, 4)
	n, err := b.transport.Receive(ctx, headerBuf)
	if err != nil {
		return err
	}
	if n < 4 {
		return fmt.Errorf("failed to read header")
	}

	header, err := ParseHeader(headerBuf)
	if err != nil {
		return err
	}

	// If no data, nothing to do
	if header.Length <= 4 {
		return nil
	}

	// 2. Read FULL packet from the beginning
	// On BNO08x I2C, starting a new read transaction restarts the packet from the beginning.
	// We read the full length we just discovered.
	fullPacket := make([]byte, header.Length)
	_, err = b.transport.Receive(ctx, fullPacket)
	if err != nil {
		return err
	}

	// Re-parse the header from the full packet to be safe
	actualHeader, _ := ParseHeader(fullPacket)
	if b.Debug {
		fmt.Printf("BNO08X: Packet Chan:%d Seq:%d Len:%d\n", actualHeader.Channel, actualHeader.SequenceNumber, actualHeader.Length)
	}

	b.handlePacket(actualHeader, fullPacket[4:])
	return nil
}

func (b *BNO08X) CheckID(ctx context.Context) error {
	b.mu.Lock()
	delete(b.readings, ReportProductIdResponse)
	b.mu.Unlock()

	// Send Product ID Request
	data := []byte{ReportProductIdRequest, 0x00}
	if err := b.sendPacket(ChanControl, data); err != nil {
		return err
	}

	// Loop for a while to find the response
	for i := 0; i < 100; i++ {
		if err := b.Process(ctx); err != nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		b.mu.Lock()
		_, ok := b.readings[ReportProductIdResponse]
		b.mu.Unlock()
		if ok {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for Product ID response")
}

func (b *BNO08X) handlePacket(header SHTPHeader, data []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.seqNumbers[header.Channel] = header.SequenceNumber

	if header.Channel == ChanInputSensorReports || header.Channel == ChanGyroRotationVector {
		// Sensor reports can be batched
		// For now, let's assume single report for simplicity or iterate if length permits
		// BNO08x reports often start with a timestamp base (0xFB or 0xFA)
		offset := 0
		for offset < len(data) {
			reportID := data[offset]
			if reportID == ReportBaseTimestamp || reportID == ReportTimestampRebase {
				if b.Debug {
					fmt.Printf("BNO08X: Skipping timestamp base 0x%02X\n", reportID)
				}
				offset += 5 // Skip timestamp base
				continue
			}

			// Get report length (this is tricky as it varies)
			// For minimal implementation, we'll focus on the target reports
			length := b.getReportLength(reportID)
			if length == 0 || offset+length > len(data) {
				if b.Debug {
					fmt.Printf("BNO08X: Unknown (0x%02X) or incomplete report at offset %d\n", reportID, offset)
				}
				break
			}

			if b.Debug {
				fmt.Printf("BNO08X: Parsing report 0x%02X\n", reportID)
			}

			res, accuracy, err := ParseSensorReport(reportID, data[offset:offset+length])
			if err == nil && res != nil {
				b.readings[reportID] = res
				b.accuracies[reportID] = accuracy
			} else if reportID == SensorReportStabilityClassifier {
				// Handle simple byte reports not in ParseSensorReport
				if len(data[offset:offset+length]) >= 5 {
					b.readings[reportID] = data[offset+4]
					b.accuracies[reportID] = int(data[offset+2] & 0x03)
				}
			}
			offset += length
		}
	} else if header.Channel == ChanControl {
		if len(data) > 0 {
			reportID := data[0]
			if reportID == ReportProductIdResponse {
				b.readings[ReportProductIdResponse] = true // Just mark as seen
			}
		}
	}
}

func (b *BNO08X) getReportLength(reportID uint8) int {
	switch reportID {
	case SensorReportAccelerometer, SensorReportLinearAcceleration, SensorReportGravity, SensorReportGyroscope, SensorReportMagnetometer:
		return 10
	case SensorReportRotationVector, SensorReportGameRotationVector:
		return 14
	case SensorReportGeomagneticRotationVector:
		return 14
	case SensorReportStabilityClassifier:
		return 6
	}
	return 0
}

func (b *BNO08X) sendPacket(channel uint8, data []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	packet := make([]byte, len(data)+4)
	header := SHTPHeader{
		Length:         uint16(len(data) + 4),
		Channel:        channel,
		SequenceNumber: b.seqNumbers[channel],
	}
	header.Encode(packet[0:4])
	copy(packet[4:], data)

	err := b.transport.Send(context.Background(), packet)
	if err == nil {
		b.seqNumbers[channel]++
	}
	return err
}

func (b *BNO08X) GetQuaternion() (Quaternion, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	q, ok := b.readings[SensorReportRotationVector].(Quaternion)
	return q, ok
}

func (b *BNO08X) GetAccelerometer() ([3]float64, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	v, ok := b.readings[SensorReportAccelerometer].([3]float64)
	return v, ok
}

func (b *BNO08X) GetGyroscope() ([3]float64, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	v, ok := b.readings[SensorReportGyroscope].([3]float64)
	return v, ok
}

func (b *BNO08X) GetMagnetometer() ([3]float64, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	v, ok := b.readings[SensorReportMagnetometer].([3]float64)
	return v, ok
}

func (b *BNO08X) GetAccuracy(reportID uint8) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.accuracies[reportID]
}

func (b *BNO08X) GetStability() (string, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	val, ok := b.readings[SensorReportStabilityClassifier].(uint8)
	if !ok {
		return "", false
	}
	classifications := []string{"Unknown", "On Table", "Stationary", "Stable", "In motion"}
	if int(val) < len(classifications) {
		return classifications[val], true
	}
	return "Unknown", true
}
