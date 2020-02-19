package base

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type SeverityHook struct{}

func (h SeverityHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level != zerolog.NoLevel {
		e.Str("severity", level.String())
	}
}

// Log 全局日志对象
var Log zerolog.Logger

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}

// Hooks 增加hooks
func Hooks(h zerolog.Hook) {

	if h == nil {
		return
	}

	Log = log.Hook(h) //SeverityHook{}

	sublogger1 := Log.With().Str("a1", "a2").Logger()
	Log.Warn().Msg("111")

	sublogger1.Debug().Msg("33222")
}

// NewSubLogger 生成子日志对象
func NewSubLogger(name string, value string) zerolog.Logger {
	return Log.With().Str(name, value).Logger()
}
