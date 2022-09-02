package saddle

import (
	"context"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/rs/zerolog"
	"github.com/saddlemc/saddle/config"
	"github.com/saddlemc/saddle/log"
	"github.com/saddlemc/saddle/plugin"
	"os"
)

func Run() {
	logger := zerolog.New(os.Stdout).
		Level(zerolog.InfoLevel).
		Output(zerolog.ConsoleWriter{
			Out:          os.Stdout,
			PartsExclude: []string{zerolog.TimestampFieldName},
		})

	chat.Global.Subscribe(chat.StdoutSubscriber{})

	cfg, err := config.Read()
	if err != nil {
		logger.Fatal().Msgf("Unable to load server config: %s", err.Error())
	}
	if cfg.Console.Debug {
		logger = logger.Level(zerolog.DebugLevel)
	}

	runPlugins, err := plugin.Initialize(logger, cfg.Plugins)
	if err != nil {
		logger.Fatal().Msgf("Error loading plugins: %v", err)
	}

	srv := server.New(&cfg.Config, log.ZerologCompat{Log: &logger})
	srv.CloseOnProgramEnd()
	if err := srv.Start(); err != nil {
		logger.Fatal().Msgf("Could not start server: %s", err.Error())
	}

	ctx, stopPlugins := context.WithCancel(context.Background())
	waitForPlugins := runPlugins(ctx)

	for srv.Accept(nil) {
		// todo: player handlers
	}

	// When the server stops, send the signal to the context to stop all running plugins.
	stopPlugins()
	// Make sure all plugins have shut down.
	waitForPlugins.Wait()
}
