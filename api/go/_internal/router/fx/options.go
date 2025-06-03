package routerfx

import (
	"github.com/huangc28/vercel-go-scaffold/api/go/_internal/router"
	"go.uber.org/fx"
)

var CoreRouterOptions = fx.Options(
	fx.Provide(
		fx.Annotate(
			router.NewRouter,
			fx.ParamTags(`group:"handlers"`),
		),
	),
)
