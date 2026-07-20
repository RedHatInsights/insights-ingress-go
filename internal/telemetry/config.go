package telemetry

type OtelConfig struct {
	Enabled              bool
	Endpoint             string
	SamplingRate         float64
	ServiceName          string
	BSPMaxQueueSize      int
	BSPMaxExportBatchSize int
	BSPScheduleDelay     int
	BSPExportTimeout     int
}
