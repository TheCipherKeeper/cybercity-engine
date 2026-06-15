// Package config отвечает за загрузку настроек движка из переменных
// окружения и CLI-флагов.
package config

import (
	"os"
	"strconv"
	"strings"
)

// EngineConfig — runtime-настройки движка.
type EngineConfig struct {
	AppName string `json:"app_name"`
	Debug   bool   `json:"debug"`

	// Engine runtime
	TickMs        int    `json:"tick_ms"`
	EngineZipURL  string `json:"engine_zip_url"`
	EngineZipPath string `json:"engine_zip_path"`

	// Kafka / Redpanda
	KafkaBootstrapServers string   `json:"kafka_bootstrap_servers"`
	KafkaGroupID          string   `json:"kafka_group_id"`
	KafkaTopics           []string `json:"kafka_topics"`

	// PostgreSQL
	DatabaseURL           string `json:"database_url"`
	SnapshotIntervalTicks int    `json:"snapshot_interval_ticks"`

	// Web/API
	Host string `json:"host"`
	Port int    `json:"port"`

	// S3 / MinIO
	S3Endpoint  string `json:"s3_endpoint"`
	S3Bucket    string `json:"s3_bucket"`
	S3AccessKey string `json:"s3_access_key"`
	S3SecretKey string `json:"s3_secret_key"`
}

// LoadEnvConfig читает переменные окружения с префиксом CYBERCITY_.
func LoadEnvConfig() EngineConfig {
	return EngineConfig{
		AppName:               envString("CYBERCITY_APP_NAME", "cybercity-engine"),
		Debug:                 envBool("CYBERCITY_DEBUG", false),
		TickMs:                envInt("CYBERCITY_TICK_MS", 1000),
		EngineZipURL:          envString("CYBERCITY_ENGINE_ZIP_URL", "http://localhost:9000/cybercity/engine.zip"),
		EngineZipPath:         envString("CYBERCITY_ENGINE_ZIP_PATH", ""),
		KafkaBootstrapServers: envString("CYBERCITY_KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"),
		KafkaGroupID:          envString("CYBERCITY_KAFKA_GROUP_ID", "cybercity-engine"),
		KafkaTopics:           envStringSlice("CYBERCITY_KAFKA_TOPICS", []string{"city.commands", "city.events"}),
		DatabaseURL:           envString("CYBERCITY_DATABASE_URL", "postgresql://engine:engine@localhost:5432/cybercity"),
		SnapshotIntervalTicks: envInt("CYBERCITY_SNAPSHOT_INTERVAL_TICKS", 10),
		Host:                  envString("CYBERCITY_HOST", "0.0.0.0"),
		Port:                  envInt("CYBERCITY_PORT", 8000),
		S3Endpoint:            envString("CYBERCITY_S3_ENDPOINT", ""),
		S3Bucket:              envString("CYBERCITY_S3_BUCKET", "cybercity"),
		S3AccessKey:           envString("CYBERCITY_S3_ACCESS_KEY", ""),
		S3SecretKey:           envString("CYBERCITY_S3_SECRET_KEY", ""),
	}
}

// ApplyFlags перезаписывает настройки из CLI-флагов.
func (c *EngineConfig) ApplyFlags(engineZip *string, host *string, port *int, debug *bool) {
	if engineZip != nil && *engineZip != "" {
		c.EngineZipURL = *engineZip
		c.EngineZipPath = *engineZip
	}
	if host != nil && *host != "" {
		c.Host = *host
	}
	if port != nil && *port != 0 {
		c.Port = *port
	}
	if debug != nil && *debug {
		c.Debug = *debug
	}
}

func envString(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return def
}

func envStringSlice(key string, def []string) []string {
	if v := os.Getenv(key); v != "" {
		return strings.Split(v, ",")
	}
	return def
}
