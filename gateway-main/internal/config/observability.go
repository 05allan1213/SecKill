package config

func applyObservability(c *Config) {
	if c == nil {
		return
	}

	if c.Observability.LogRotation.MaxSizeMB <= 0 {
		c.Observability.LogRotation.MaxSizeMB = 100
	}
	if c.Observability.LogRotation.MaxBackups <= 0 {
		c.Observability.LogRotation.MaxBackups = 7
	}
	if c.Observability.AccessLog.SummaryMaxBytes <= 0 {
		c.Observability.AccessLog.SummaryMaxBytes = 128
	}

	if !c.Observability.Trace.Enabled {
		c.Middlewares.Trace = false
		c.Telemetry.Name = ""
		c.Telemetry.Endpoint = ""
		c.Telemetry.Batcher = ""
		c.Telemetry.Sampler = 0
		return
	}

	c.Middlewares.Trace = true
	if c.Telemetry.Sampler <= 0 {
		c.Telemetry.Sampler = 0.01
	}
}
