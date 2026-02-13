package bno08x

import (
	"context"
	"fmt"
	"sync"
)

type BNO08X struct {
	transport Transport
	mu        sync.Mutex

	seqNumbers [6]uint8
	readings   map[uint8]interface{}
	accuracies map[uint8]int
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
	// 1. Read header (4 bytes)
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

	if header.Length <= 4 {
		return nil // No data
	}

	// 2. Read remaining cargo
	cargoLen := int(header.Length) - 4
	cargo := make([]byte, cargoLen)
	_, err = b.transport.Receive(ctx, cargo)
	if err != nil {
		return err
	}

	b.handlePacket(header, cargo)
	return nil
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
				offset += 5 // Skip timestamp base
				continue
			}

			// Get report length (this is tricky as it varies)
			// For minimal implementation, we'll focus on the target reports
			length := b.getReportLength(reportID)
			if length == 0 || offset+length > len(data) {
				break
			}

			res, accuracy, err := ParseSensorReport(reportID, data[offset:offset+length])
			if err == nil && res != nil {
				b.readings[reportID] = res
				b.accuracies[reportID] = accuracy
			}
			offset += length
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
