package config

import "os"

type OTelConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	ServiceName    string `mapstructure:"service_name"`
	Endpoint      string `mapstructure:"endpoint"`
	Insecure      bool   `mapstructure:"insecure"`
	TraceEnabled  bool   `mapstructure:"trace_enabled"`
	MetricEnabled bool   `mapstructure:"metric_enabled"`
}

func (c *Config) GetOTelConfig() OTelConfig {
	return OTelConfig{
		Enabled:        os.Getenv("OTEL_ENABLED") == "true",
		ServiceName:    getEnv("OTEL_SERVICE_NAME", "hermes-server"),
		Endpoint:       getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		Insecure:       getEnv("OTEL_INSECURE", "true") == "true",
		TraceEnabled:   getEnv("OTEL_TRACE_ENABLED", "true") == "true",
		MetricEnabled: getEnv("OTEL_METRIC_ENABLED", "false") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
