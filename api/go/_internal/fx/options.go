package appfx

import (
	"github.com/huangc28/kikichoice-be/api/go/_internal/configs"
	"github.com/huangc28/kikichoice-be/api/go/_internal/db"
	"github.com/huangc28/kikichoice-be/api/go/_internal/pkg/logger"
	"go.uber.org/fx"
)

var CoreConfigOptions = fx.Provide(
	logger.NewLogger,
	configs.NewViper,
	configs.NewConfig,
	db.NewSQLXPostgresDB, // sql concrete instance
	fx.Annotate(
		db.NewSQLXPostgresDB,
		fx.As(new(db.Conn)),
	), // sql interface instance
)
