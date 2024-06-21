package main

import (
	"arimadj-helper/internal/bot"
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	tgapi, err := tgbotapi.NewBotAPI("7447550770:AAHaO6tRmqNtb53fD9cIXPJVjYi1mHN3i_0")
	if err != nil {
		log.Panic(err)
	}

	logrus.Info("tgapi")
	bot := bot.New(tgapi)
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	logrus.Info("run")
	bot.Run(ctx)
	//server := NewServer()
	//server.Run()
	logrus.Info("ctx")
	<-ctx.Done()

}
