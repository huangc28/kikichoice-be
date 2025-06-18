package logger

import (
	"github.com/huangc28/kikichoice-be/api/go/_internal/configs"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TagLogger(tag string) fx.Option {
	return fx.Supply(
		fx.Annotate(
			tag,
			fx.ResultTags(`name:"service-name"`),
		),
	)
}

type NewLoggerParams struct {
	fx.In

	Cfg         *configs.Config
	ServiceName string `name:"service-name"`
}

func NewLogger(p NewLoggerParams) *zap.SugaredLogger {
	// If log service name is empty, give it a default name.
	if p.ServiceName == "" {
		p.ServiceName = "default"
	}

	var (
		baseLogger *zap.Logger
		err        error
	)

	logConfig := zap.NewProductionConfig()
	if p.Cfg.ENV == configs.Dev {
		logConfig = zap.NewDevelopmentConfig()
	}

	logConfig.InitialFields = map[string]any{
		"service": p.ServiceName,
	}

	baseLogger, err = logConfig.Build()
	if err != nil {
		panic(err)
	}

	var cores []zapcore.Core
	cores = append(cores, baseLogger.Core())

	combinedCore := zapcore.NewTee(cores...)
	baseLogger = zap.New(combinedCore, zap.WithCaller(true))

	return baseLogger.Sugar()
}
