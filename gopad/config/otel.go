package config

import (
	"fmt"
)

type OtelConfig struct {
	InstanceID string       `toml:"instance_id"`
	Trace      *TraceConfig `toml:"trace"`
}

func (c OtelConfig) String() string {
	return fmt.Sprintf("\n  InstanceID: %s\n  Trace: %s",
		c.InstanceID,
		c.Trace,
	)
}

type TraceConfig struct {
	Endpoint string `toml:"endpoint"`
	Insecure bool   `toml:"insecure"`
}

func (c TraceConfig) String() string {
	return fmt.Sprintf("\n   Endpoint: %s\n   Insecure: %t",
		c.Endpoint,
		c.Insecure,
	)
}
