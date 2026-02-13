package bno08x

// SHTP Channels
const (
	ChanSHTPCommand        = 0
	ChanExecutable         = 1
	ChanControl            = 2
	ChanInputSensorReports = 3
	ChanWakeSensorReports  = 4
	ChanGyroRotationVector = 5
)

// SH-2 Report IDs
const (
	ReportGetFeatureRequest  = 0xFE
	ReportSetFeatureCommand  = 0xFD
	ReportGetFeatureResponse = 0xFC
	ReportBaseTimestamp      = 0xFB
	ReportTimestampRebase    = 0xFA
	ReportProductIdResponse  = 0xF8
	ReportProductIdRequest   = 0xF9
	ReportCommandRequest     = 0xF2
	ReportCommandResponse    = 0xF1
)

// Sensor Report IDs
const (
	SensorReportAccelerometer             = 0x01
	SensorReportGyroscope                 = 0x02
	SensorReportMagnetometer              = 0x03
	SensorReportLinearAcceleration        = 0x04
	SensorReportRotationVector            = 0x05
	SensorReportGravity                   = 0x06
	SensorReportGameRotationVector        = 0x08
	SensorReportGeomagneticRotationVector = 0x09
	SensorReportStepCounter               = 0x11
	SensorReportRawAccelerometer          = 0x14
	SensorReportRawGyroscope              = 0x15
	SensorReportRawMagnetometer           = 0x16
	SensorReportShakeDetector             = 0x19
	SensorReportStabilityClassifier       = 0x13
	SensorReportActivityClassifier        = 0x1E
	SensorReportGyroIntegratedRV          = 0x2A
)

// Q-Point constants for scaling
const (
	QPoint14 = 14
	QPoint12 = 12
	QPoint10 = 10
	QPoint9  = 9
	QPoint8  = 8
	QPoint4  = 4
)
