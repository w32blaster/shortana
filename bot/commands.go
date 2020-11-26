package bot

import (
	"github.com/w32blaster/shortana/db"
	"github.com/w32blaster/shortana/stats"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Command struct {
	db       *db.Database
	bot      *tgbotapi.BotAPI
	hostname string
	stats    *stats.Statistics
}

func (c Command) NotAllowedToSpeak(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	sendMsg(c.bot, chatID, "Sorry, "+message.From.UserName+", I can't talk to you")
}

// ProcessCommands acts when user sent to a bot some command, for example "/command arg1 arg2"
func (c Command) ProcessCommands(message *tgbotapi.Message) {

	chatID := message.Chat.ID
	command := extractCommand(message.Command())
	log.Println("This is command /" + command)

	switch command {

	case "start":
		sendMsg(c.bot, chatID, "Yay! Welcome! I can work with Shortana and manage your shortened URLs pretty easy")

	case "list":
		renderShortenedURLsList(c.bot, chatID, c.db, c.hostname)

	default:
		sendMsg(c.bot, chatID, "Sorry, I don't recognyze such command: "+command+", please call /help to get full list of commands I understand")
	}
}

func (c Command) ProcessButtonCallback(callbackQuery *tgbotapi.CallbackQuery) {

}

func renderShortenedURLsList(bot *tgbotapi.BotAPI, chatID int64, database *db.Database, hostname string) {
	shortenedUrls, err := database.GetAll()
	if err != nil {
		sendMsg(bot, chatID, "something wrong happened")
		log.Println("Error reading the database, " + err.Error())
	}
	if len(shortenedUrls) == 0 {
		sendMsg(bot, chatID, "no short URLs yet. You can add a new one by hitting /add command")
	}
	var sb strings.Builder
	for i, k := range shortenedUrls {
		sb.WriteString(strconv.Itoa(i + 1))
		sb.WriteString(") [")
		sb.WriteString(hostname)
		sb.WriteString("/")
		sb.WriteString(k.ShortUrl)
		sb.WriteString("](")
		sb.WriteString(k.FullUrl)
		sb.WriteString(")")
		sb.WriteString("\n")
	}
	sendMsg(bot, chatID, sb.String())
}

// properly extracts command from the input string, removing all unnecessary parts
// please refer to unit tests for details
func extractCommand(rawCommand string) string {

	command := rawCommand

	// remove slash if necessary
	if rawCommand[0] == '/' {
		command = command[1:]
	}

	// if command contains the name of our bot, remote it
	command = strings.Split(command, "@")[0]
	command = strings.Split(command, " ")[0]

	return command
}

// simply send a message to bot in Markdown format
func sendMsg(bot *tgbotapi.BotAPI, chatID int64, textMarkdown string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, textMarkdown)
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true

	// send the message
	resp, err := bot.Send(msg)
	if err != nil {
		log.Println("bot.Send:", err, resp)
		return resp, err
	}

	return resp, err
}
