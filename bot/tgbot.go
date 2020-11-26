package bot

import (
	"github.com/w32blaster/shortana/db"
	"log"
	"net/http"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func Start(database *db.Database, botToken string, port, acceptFromUser int, hostname string, isDebug bool) {

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		panic("Bot doesn't work. Reason: " + err.Error())
	}

	bot.Debug = isDebug

	cmd := Command{
		db:       database,
		bot:      bot,
		hostname: hostname,
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)
	updates := bot.ListenForWebhook("/" + botToken)

	go http.ListenAndServe(":"+strconv.Itoa(port), nil)

	for update := range updates {

		if update.Message != nil {

			if !isUserAllowedToSpeakToBot(update.Message.From.ID, acceptFromUser) {

				cmd.NotAllowedToSpeak(update.Message)

			} else if update.Message.IsCommand() {

				// This is a command starting with slash
				cmd.ProcessCommands(update.Message)

			} else {

				if update.Message.ReplyToMessage == nil {

					// This is a simple text
					log.Println("This is plain text: " + update.Message.Text)
					//commands.ProcessSimpleText(bot, update.Message)

				}
			}

		} else if update.CallbackQuery != nil {

			if isUserAllowedToSpeakToBot(update.CallbackQuery.From.ID, acceptFromUser) {

				// this is the callback after a button click
				cmd.ProcessButtonCallback(update.CallbackQuery)
			}

		}

	}

}

func isUserAllowedToSpeakToBot(realUser, expectedUser int) bool {
	if expectedUser == 0 {
		return true // user is not set, assume we can speak to everyone
	}
	return realUser == expectedUser
}
