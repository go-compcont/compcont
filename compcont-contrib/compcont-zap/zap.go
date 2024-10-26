package compcontzap

import (
	"net/url"
	"strconv"

	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ExtraConfig struct {
	Level             *string  `ccf:"level"`
	DisableCaller     *bool    `ccf:"disable_caller"`
	DisableStacktrace *bool    `ccf:"disable_stacktrace"`
	Encoding          *string  `ccf:"encoding"`
	OutputPaths       []string `ccf:"output_paths"`
	ErrorOutputPaths  []string `ccf:"error_output_paths"`
}

func (c *ExtraConfig) MergeTo(input zap.Config) (output zap.Config, err error) {
	output = input
	if c.Level != nil {
		output.Level, err = zap.ParseAtomicLevel(*c.Level)
		if err != nil {
			return
		}
	}
	if c.DisableCaller != nil {
		output.DisableCaller = *c.DisableCaller
	}
	if c.DisableStacktrace != nil {
		output.DisableStacktrace = *c.DisableStacktrace
	}
	if c.Encoding != nil {
		output.Encoding = *c.Encoding
	}
	if len(c.OutputPaths) > 0 {
		output.OutputPaths = c.OutputPaths
	}
	if len(c.ErrorOutputPaths) > 0 {
		output.ErrorOutputPaths = c.ErrorOutputPaths
	}
	return
}

type Config struct {
	BaseConfig  string      `ccf:"base_config"` // "","development","production"
	ExtraConfig ExtraConfig `ccf:"extra_config"`
}

func New(cfg Config) (c *zap.Logger, err error) {
	var baseCfg zap.Config
	switch cfg.BaseConfig {
	case "development":
		baseCfg = zap.NewDevelopmentConfig()
	case "production":
		baseCfg = zap.NewProductionConfig()
	case "":
	default:
		panic("unknown logger base config: " + cfg.BaseConfig)
	}
	finalCfg, err := cfg.ExtraConfig.MergeTo(baseCfg)
	if err != nil {
		return
	}
	c, err = finalCfg.Build()
	if err != nil {
		return
	}
	return
}

type sinkFunc struct {
	CloseFunc func() error
	SyncFunc  func() error
	WriteFunc func(p []byte) (n int, err error)
}

func (s *sinkFunc) Close() error {
	return s.CloseFunc()
}
func (s *sinkFunc) Sync() error {
	return s.SyncFunc()
}
func (s *sinkFunc) Write(p []byte) (n int, err error) {
	return s.WriteFunc(p)
}

func init() {
	err := zap.RegisterSink("lumberjack", func(u *url.URL) (sink zap.Sink, err error) {
		q := u.Query()
		parseQueryInt := func(key string) (int, error) {
			str := q.Get(key)
			if str == "" {
				return 0, nil
			}
			ret, err := strconv.ParseInt(str, 10, 64)
			if err != nil {
				return 0, err
			}
			return int(ret), nil
		}
		parseQueryBool := func(key string) (bool, error) {
			str := q.Get(key)
			if str == "" {
				return false, nil
			}
			ret, err := strconv.ParseBool(str)
			if err != nil {
				return false, err
			}
			return ret, err
		}

		l := &lumberjack.Logger{}

		switch u.Hostname() {
		case "relative-path":
			l.Filename = "." + u.Path
		case "absolute-path":
			l.Filename = u.Path
		default:
			panic("unknown hostname")
		}
		if l.MaxSize, err = parseQueryInt("max_size"); err != nil {
			return
		}
		if l.MaxBackups, err = parseQueryInt("max_backups"); err != nil {
			return
		}
		if l.MaxAge, err = parseQueryInt("max_age"); err != nil {
			return
		}
		if l.Compress, err = parseQueryBool("compress"); err != nil {
			return
		}
		if l.LocalTime, err = parseQueryBool("local_time"); err != nil {
			return
		}

		sink = &sinkFunc{
			CloseFunc: l.Close,
			WriteFunc: l.Write,
			SyncFunc: func() error {
				return nil
			},
		}
		return
	})
	if err != nil {
		panic(err)
	}
}
