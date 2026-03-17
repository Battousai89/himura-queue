package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server   ServerConfig
	Queue    QueueConfig
	Worker   WorkerConfig
	Snapshot SnapshotConfig
}

type ServerConfig struct {
	TCPPort  int
	HTTPPort int
}

type QueueConfig struct {
	ShardCount int
}

type WorkerConfig struct {
	MinWorkers  int
	MaxWorkers  int
	IdleTimeout int
}

type SnapshotConfig struct {
	Path     string
	Interval int
}

func Default() *Config {
	return &Config{
		Server: ServerConfig{
			TCPPort:  9000,
			HTTPPort: 9001,
		},
		Queue: QueueConfig{
			ShardCount: 8,
		},
		Worker: WorkerConfig{
			MinWorkers:  4,
			MaxWorkers:  100,
			IdleTimeout: 30,
		},
		Snapshot: SnapshotConfig{
			Path:     "data/snapshot.bin",
			Interval: 30,
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var section string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = line[1 : len(line)-1]
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(strings.Trim(parts[1], `"`))
		value = strings.Trim(value, `"`)

		switch section {
		case "server":
			switch key {
			case "tcp_port":
				cfg.Server.TCPPort, _ = strconv.Atoi(value)
			case "http_port":
				cfg.Server.HTTPPort, _ = strconv.Atoi(value)
			}
		case "queue":
			switch key {
			case "shard_count":
				cfg.Queue.ShardCount, _ = strconv.Atoi(value)
			}
		case "worker":
			switch key {
			case "min_workers":
				cfg.Worker.MinWorkers, _ = strconv.Atoi(value)
			case "max_workers":
				cfg.Worker.MaxWorkers, _ = strconv.Atoi(value)
			case "idle_timeout_sec":
				cfg.Worker.IdleTimeout, _ = strconv.Atoi(value)
			}
		case "snapshot":
			switch key {
			case "path":
				cfg.Snapshot.Path = value
			case "interval_sec":
				cfg.Snapshot.Interval, _ = strconv.Atoi(value)
			}
		}
	}

	return cfg, scanner.Err()
}

func (c *Config) SnapshotInterval() time.Duration {
	return time.Duration(c.Snapshot.Interval) * time.Second
}
