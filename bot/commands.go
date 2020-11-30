package bot

import (
	"log"
	"strconv"
	"strings"

	"github.com/w32blaster/shortana/db"
	"github.com/w32blaster/shortana/geoip"
	"github.com/w32blaster/shortana/stats"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type addingStep int

const (
	None = 1 + iota
	RequestedShortenedUrl
	RequestedTargetLink
	RequestedDescription
	RequestedButtonIsPrivateOrPublic

	public = "p"
)

type Command struct {
	db               *db.Database
	bot              *tgbotapi.BotAPI
	hostname         string
	stats            *stats.Statistics
	geoIP            *geoip.GeoIP
	step             addingStep // when we start a dialog to add a new data, we should remember step for it
	halfSavedShortID string     // short URL saved in DB with half filled data in it
}

func (c *Command) NotAllowedToSpeak(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	sendMsg(c.bot, chatID, "Sorry, "+message.From.UserName+", I can't talk to you")
}

// ProcessCommands acts when user sent to a bot some command, for example "/command arg1 arg2"
func (c *Command) ProcessCommands(message *tgbotapi.Message) {

	chatID := message.Chat.ID
	command := extractCommand(message.Command())
	log.Println("This is command /" + command)

	switch command {

	case "start":
		sendMsg(c.bot, chatID, "Yay! Welcome! I can work with Shortana and manage your shortened URLs pretty easy")

	case "list":
		renderShortenedURLsList(c.bot, chatID, c.db, c.hostname)

	case "add":
		c.initiateAdding(chatID)

	case "stats":
		c.printStatisticTemp(chatID)

	case "download":
		fnOnUpdate := func(msg string) {
			msg = strings.Replace(msg, "/", " ", -1)
			sendMsg(c.bot, chatID, msg)
		}
		if err := c.geoIP.DownloadGeoIPDatabase(fnOnUpdate); err != nil {
			sendMsg(c.bot, chatID, "Database update failed, reason is "+err.Error())
		}
		sendMsg(c.bot, chatID, "Yay! Database was updated properly")

	default:
		sendMsg(c.bot, chatID, "Sorry, I don't recognyze such command: "+command+", please call /help to get full list of commands I understand")
	}
}

// initiate process of adding a new short URL, this process consists of few steps, so we keep state inside
// internal variable "step"
func (c *Command) initiateAdding(chatID int64) {
	c.step = RequestedShortenedUrl
	sendMsg(c.bot, chatID, "Ok, can you send me the short url please? Send me just suffix without a hostname")
}

func (c *Command) ProcessSimpleText(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	switch c.step {

	// Step 1: requested the Short URL
	case RequestedShortenedUrl:
		if err := c.db.SaveShortUrlObject(&db.ShortURL{ShortUrl: message.Text, IsPublic: false}); err != nil {
			c.step = None
			sendMsg(c.bot, chatID, "Sorry, I tried to save your short URL but failed. "+
				"Can you start from beginning please?")
			log.Println("Failed to save a new short link, error " + err.Error())
			return
		}
		c.halfSavedShortID = message.Text
		c.step = RequestedTargetLink
		sendMsg(c.bot, chatID, "Ok, can you send me the full target URL where the short URL will lead to?")

	// step 2: a new short URL is saved with only one value and a target link was requested
	case RequestedTargetLink:

		if len(message.Text) == 0 || !strings.HasPrefix(message.Text, "http") {
			sendMsg(c.bot, chatID, "Incorrect value, please send me the full URL link, please")
			return
		}

		if err := c.db.UpdateShortUrl(c.halfSavedShortID, "TargetUrl", message.Text); err != nil {
			sendMsg(c.bot, chatID, "Sorry, can't update TargetURL "+
				"Can you send me once again please?")
			log.Println("Failed to update a short link, error " + err.Error())
			return
		}

		c.step = RequestedDescription
		sendMsg(c.bot, chatID, "Nice one. Now send me the description, please?")

	// step 3: description
	case RequestedDescription:
		if err := c.db.UpdateShortUrl(c.halfSavedShortID, "Description", message.Text); err != nil {
			sendMsg(c.bot, chatID, "Sorry, can't update Description "+
				"Can you send me once again please?")
			log.Println("Failed to update a short link, error " + err.Error())
			return
		}

		c.step = RequestedButtonIsPrivateOrPublic
		resp, _ := sendMsg(c.bot, chatID, "Nice one. Is it public or private?")
		renderPublicPrivateButtons(c.bot, chatID, resp.MessageID)
	}
}

func (c *Command) ProcessButtonCallback(callbackQuery *tgbotapi.CallbackQuery) {

	// notify the telegram that we processed the button, it will turn "loading indicator" off
	defer c.bot.AnswerCallbackQuery(tgbotapi.CallbackConfig{
		CallbackQueryID: callbackQuery.ID,
	})

	if c.step == RequestedButtonIsPrivateOrPublic {

		if err := c.db.UpdateShortUrl(c.halfSavedShortID, "IsPublic", callbackQuery.Data == public); err != nil {
			sendMsg(c.bot, callbackQuery.Message.Chat.ID, "Sorry, can't update short link "+
				"Can you send me once again please?")
			log.Println("Failed to update a short link, error " + err.Error())
			return
		}

		// here, a new short link is ready to be used
		c.step = None
		c.halfSavedShortID = ""

		// delete buttons
		msg := tgbotapi.NewDeleteMessage(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID)
		c.bot.Send(msg)

		sendMsg(c.bot, callbackQuery.Message.Chat.ID, "New Short URL is saved")
	}
}

func (c *Command) printStatisticTemp(chatID int64) {
	var sb strings.Builder
	stats, err := c.db.GetAllStatistics()
	if err != nil {
		sendMsg(c.bot, chatID, "Error getting stats: "+err.Error())
		return
	}

	for _, k := range stats {
		sb.WriteString("> URL: ")
		sb.WriteString(k.ShortUrl)
		sb.WriteString(", country: ")
		sb.WriteString(k.CountryCode)
		sb.WriteString(", city: ")
		sb.WriteString(k.City)
		sb.WriteString(", IP: ")
		sb.WriteString(k.UserIpAddress)
		sb.WriteString("\n---\n")
	}
	sendMsg(c.bot, chatID, sb.String())
}

func renderPublicPrivateButtons(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
	buttonRows := make([][]tgbotapi.InlineKeyboardButton, 1)

	buttonRows[0] = []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Pubic", public),
		tgbotapi.NewInlineKeyboardButtonData("Private", "priv"),
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
	keyboardMsg := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, keyboard)
	bot.Send(keyboardMsg)
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
		sb.WriteString(k.TargetUrl)
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
