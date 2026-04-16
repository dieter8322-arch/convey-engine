// SPDX-FileCopyrightText: 2026 dieter8322-arch
// SPDX-License-Identifier: MIT

package config

import "os"

type Config struct {
	HTTP          HTTPConfig
	Database      DatabaseConfig
	Temporal      TemporalConfig
	ObjectStorage ObjectStorageConfig
}

type HTTPConfig struct {
	Addr string
}

type DatabaseConfig struct {
	DSN string
}

type TemporalConfig struct {
	Address   string
	Namespace string
	TaskQueue string
}

type ObjectStorageConfig struct {
	LocalRoot string
}

func Load() Config {
	return Config{
		HTTP: HTTPConfig{
			Addr: getEnv("HTTP_ADDR", ":8080"),
		},
		Database: DatabaseConfig{
			DSN: getEnv("DB_DSN", "postgres://convey:convey@localhost:5432/convey_app?sslmode=disable"),
		},
		Temporal: TemporalConfig{
			Address:   getEnv("TEMPORAL_ADDRESS", "127.0.0.1:7233"),
			Namespace: getEnv("TEMPORAL_NAMESPACE", "default"),
			TaskQueue: getEnv("TEMPORAL_TASK_QUEUE", "pipeline-runs"),
		},
		ObjectStorage: ObjectStorageConfig{
			LocalRoot: getEnv("OBJECT_STORAGE_LOCAL_ROOT", ".data/objects"),
		},
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
