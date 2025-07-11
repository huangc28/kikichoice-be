package commands

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/fx"
)

type BotCommand string

func (c BotCommand) String() string {
	return string(c)
}

var (
	AddProduct BotCommand = "add"
)

type CommandHandler interface {
	Handle(msg *tgbotapi.Message) error
	HandleReply(ctx context.Context, msg *tgbotapi.Message) error
	Command() BotCommand
}

func AsCommandHandler(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(CommandHandler)),
		fx.ResultTags(`group:"command_handlers"`),
	)
}

func NewCommandHandlerMap(handlers []CommandHandler) map[BotCommand]CommandHandler {
	handlerMap := make(map[BotCommand]CommandHandler)
	for _, handler := range handlers {
		handlerMap[handler.Command()] = handler
	}
	return handlerMap
}
