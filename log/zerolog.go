package log

import "github.com/rs/zerolog"

// ZerologCompat is a zerolog compatibility layer for dragonfly.
type ZerologCompat struct {
	Log *zerolog.Logger
}

func (z ZerologCompat) Warnf(format string, args ...any) {
	z.Log.Warn().Msgf(format, args...)
}

func (z ZerologCompat) Errorf(format string, a ...any) {
	z.Log.Error().Msgf(format, a...)
}

func (z ZerologCompat) Debugf(format string, a ...any) {
	z.Log.Debug().Msgf(format, a...)
}

func (z ZerologCompat) Infof(format string, v ...any) {
	z.Log.Info().Msgf(format, v...)
}

func (z ZerologCompat) Fatalf(format string, v ...any) {
	z.Log.Fatal().Msgf(format, v...)
}
