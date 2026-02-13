package bno08x

import (
	"encoding/binary"
	"math"
)

// SensorData represents the parsed data from a sensor report
type SensorData struct {
	Accelerometer [3]float64
	Gyroscope     [3]float64
	Magnetometer  [3]float64
	Rotation      Quaternion
	Accuracy      int
}

// Quaternion represents a 4D rotation vector
type Quaternion struct {
	I, J, K, Real float64
}

// ParseSensorReport parses a single sensor report from a buffer
func ParseSensorReport(reportID uint8, data []byte) (interface{}, int, error) {
	switch reportID {
	case SensorReportRotationVector, SensorReportGameRotationVector, SensorReportGeomagneticRotationVector:
		q, accuracy := parseQuaternion(data, getScalar(reportID))
		return q, accuracy, nil
	case SensorReportAccelerometer, SensorReportLinearAcceleration, SensorReportGravity:
		v, accuracy := parseVector3(data, getScalar(reportID))
		return v, accuracy, nil
	case SensorReportGyroscope:
		v, accuracy := parseVector3(data, getScalar(reportID))
		return v, accuracy, nil
	case SensorReportMagnetometer:
		v, accuracy := parseVector3(data, getScalar(reportID))
		return v, accuracy, nil
	}
	return nil, 0, nil
}

func parseQuaternion(data []byte, scalar float64) (Quaternion, int) {
	// Offset 4: i, 6: j, 8: k, 10: real, 12: accuracy
	i := float64(int16(binary.LittleEndian.Uint16(data[4:6]))) * scalar
	j := float64(int16(binary.LittleEndian.Uint16(data[6:8]))) * scalar
	k := float64(int16(binary.LittleEndian.Uint16(data[8:10]))) * scalar
	real := float64(int16(binary.LittleEndian.Uint16(data[10:12]))) * scalar
	accuracy := int(data[12] & 0x03)
	return Quaternion{I: i, J: j, K: k, Real: real}, accuracy
}

func parseVector3(data []byte, scalar float64) ([3]float64, int) {
	// Offset 4: x, 6: y, 8: z
	x := float64(int16(binary.LittleEndian.Uint16(data[4:6]))) * scalar
	y := float64(int16(binary.LittleEndian.Uint16(data[6:8]))) * scalar
	z := float64(int16(binary.LittleEndian.Uint16(data[8:10]))) * scalar
	accuracy := int(data[10] & 0x03)
	return [3]float64{x, y, z}, accuracy
}

func getScalar(reportID uint8) float64 {
	switch reportID {
	case SensorReportRotationVector, SensorReportGameRotationVector:
		return math.Pow(2, -14)
	case SensorReportGeomagneticRotationVector:
		return math.Pow(2, -12)
	case SensorReportAccelerometer, SensorReportLinearAcceleration, SensorReportGravity:
		return math.Pow(2, -8)
	case SensorReportGyroscope:
		return math.Pow(2, -9)
	case SensorReportMagnetometer:
		return math.Pow(2, -4)
	}
	return 1.0
}

// ToEuler converts a quaternion to Euler angles in degrees (Pitch, Roll, Yaw)
func (q Quaternion) ToEuler() (float64, float64, float64) {
	i, j, k, real := q.I, q.J, q.K, q.Real

	roll := math.Atan2(2*(real*i+j*k), 1-2*(i*i+j*j))

	sinp := 2 * (real*j - k*i)
	var pitch float64
	if math.Abs(sinp) >= 1 {
		pitch = math.Copysign(math.Pi/2, sinp)
	} else {
		pitch = math.Asin(sinp)
	}

	yaw := math.Atan2(2*(real*k+i*j), 1-2*(j*j+k*k))

	return pitch * 180 / math.Pi, roll * 180 / math.Pi, yaw * 180 / math.Pi
}
