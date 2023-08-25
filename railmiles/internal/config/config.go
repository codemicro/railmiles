package config

import (
	"fmt"
	"git.tdpain.net/pkg/cfger"
	"github.com/codemicro/railmiles/railmiles/internal/util"
)

type Config struct {
	Debug bool
	HTTP  struct {
		Host string
		Port int
	}
	RealTimeTrains struct {
		Username string
		Password string
	}
	Database struct {
		DSN string
	}
}

func (c *Config) HTTPAddress() string {
	return fmt.Sprintf("%s:%d", c.HTTP.Host, c.HTTP.Port)
}

func Load() (*Config, error) {
	cl := cfger.New()
	const configFilename = "config.yml"
	if err := cl.Load(configFilename); err != nil {
		return nil, util.Wrap(err, "loading config from %s", configFilename)
	}

	conf := new(Config)

	conf.Debug = cl.WithDefault("debug", false).AsBool()

	conf.HTTP.Host = cl.WithDefault("http.host", "127.0.0.1").AsString()
	conf.HTTP.Port = cl.WithDefault("http.port", 8080).AsInt()

	conf.RealTimeTrains.Username = cl.Required("realtimetrains.username").AsString()
	conf.RealTimeTrains.Password = cl.Required("realtimetrains.password").AsString()

	conf.Database.DSN = cl.WithDefault("database.dsn", "railmiles.db").AsString()

	return conf, nil
}
