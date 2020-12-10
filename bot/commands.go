package bot

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

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

	public                = "p"
	ButtonDeleteMsgPrefix = "dM" // for button "delete message"
	ButtonUpdateMsgPrefix = "uM" // for button "update message"
	Separator             = "#"
)

var (
	patternCommandStatsForURL       = regexp.MustCompile(`^(stats)(\d+)$`)
	patternCommandStatsForURLOneDay = regexp.MustCompile(`^stats(\d+)x(\d{8})$`)
	patternCommandStatsOneView      = regexp.MustCompile(`^stats(\d+)x(\d{8})x(\d+)$`)
	patternCommandStatsView         = regexp.MustCompile(`^(view)(\d+)$`)

	funcMap = template.FuncMap{
		"markdownEscape": markdownEscape,
		"formatDate": func(dateTime time.Time) string {
			return markdownEscape(dateTime.Format(time.RFC822))
		},
	}
)

type (
	StatsForURLOneDay struct {
		Views        []db.OneViewStatistic
		ShortURL     db.ShortURL
		SelectedDate string
	}

	StatsGroupedByURLData struct {
		Stats    map[string]db.OneURLSummaryStatistics
		Hostname string
	}

	StatsForOneURL struct {
		Stats    map[string]db.OneDaySummaryStatistics
		ShortURL db.ShortURL
	}

	Command struct {
		db               *db.Database
		bot              *tgbotapi.BotAPI
		hostname         string
		stats            *stats.Statistics
		geoIP            *geoip.GeoIP
		step             addingStep // when we start a dialog to add a new data, we should remember step for it
		halfSavedShortID string     // short URL saved in DB with half filled data in it
	}
)

func (c *Command) NotAllowedToSpeak(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	sendEscMsg(c.bot, chatID, "Sorry, "+message.From.UserName+", I can't talk to you")
}

// ProcessCommands acts when user sent to a bot some command, for example "/command arg1 arg2"
func (c *Command) ProcessCommands(message *tgbotapi.Message) {

	chatID := message.Chat.ID
	command := extractCommand(message.Command())
	log.Println("This is command /" + command)

	switch command {

	case "start":
		sendEscMsg(c.bot, chatID, "Yay! Welcome! I can work with Shortana and manage your shortened URLs pretty easy")

	case "list":
		renderShortenedURLsList(c.bot, chatID, c.db, c.hostname)

	case "add":
		c.initiateAdding(chatID)

	case "download":
		fnOnUpdate := func(msg string) {
			sendEscMsg(c.bot, chatID, msg)
		}
		if err := c.geoIP.DownloadGeoIPDatabase(fnOnUpdate); err != nil {
			sendEscMsg(c.bot, chatID, "Database update failed, reason is "+err.Error())
		}
		sendEscMsg(c.bot, chatID, "Yay! Database was updated properly")

	default:
		if strings.HasPrefix(command, "stats") {
			c.renderStats(command, chatID, "")
			return
		}

		if patternCommandStatsView.MatchString(command) {

			// print data for one single visit
			c.getStatisticOneView(chatID, command)
			return
		}

		sendEscMsg(c.bot, chatID, "Sorry, I don't recognyze such command: "+command+", please call /help to get full list of commands I understand")

	}
}

func (c *Command) renderStats(command string, chatID int64, messageIDtoReplace string) {

	intMessageID := 0
	if len(messageIDtoReplace) > 0 {
		var err error
		intMessageID, err = strconv.Atoi(messageIDtoReplace)
		if err != nil {
			sendMsg(c.bot, chatID, "Error parsing message ID")
		}
	}

	if command == "stats" {

		// print statistic summary per all URLs
		c.printStatisticSummary(chatID, intMessageID)
	} else if patternCommandStatsForURL.MatchString(command) {

		// print statistics for one ShortURL only
		c.getStatisticOneURL(chatID, command, intMessageID)
	} else if patternCommandStatsForURLOneDay.MatchString(command) {

		// print statistics for one Short URL for one specific day
		c.getStatisticOneURLOneDay(chatID, command, intMessageID)
	}
}

// initiate process of adding a new short URL, this process consists of few steps, so we keep state inside
// internal variable "step"
func (c *Command) initiateAdding(chatID int64) {
	c.step = RequestedShortenedUrl
	sendEscMsg(c.bot, chatID, "Ok, can you send me the short url please? Send me just suffix without a hostname")
}

func (c *Command) ProcessSimpleText(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	switch c.step {

	// Step 1: requested the Short URL
	case RequestedShortenedUrl:
		if err := c.db.SaveShortUrlObject(&db.ShortURL{ShortUrl: message.Text, IsPublic: false}); err != nil {
			c.step = None
			sendEscMsg(c.bot, chatID, "Sorry, I tried to save your short URL but failed. "+
				"Can you start from beginning please?")
			log.Println("Failed to save a new short link, error " + err.Error())
			return
		}
		c.halfSavedShortID = message.Text
		c.step = RequestedTargetLink
		sendEscMsg(c.bot, chatID, "Ok, can you send me the full target URL where the short URL will lead to?")

	// step 2: a new short URL is saved with only one value and a target link was requested
	case RequestedTargetLink:

		if len(message.Text) == 0 || !strings.HasPrefix(message.Text, "http") {
			sendEscMsg(c.bot, chatID, "Incorrect value, please send me the full URL link, please")
			return
		}

		if err := c.db.UpdateShortUrl(c.halfSavedShortID, "TargetUrl", message.Text); err != nil {
			sendEscMsg(c.bot, chatID, "Sorry, can't update TargetURL "+
				"Can you send me once again please?")
			log.Println("Failed to update a short link, error " + err.Error())
			return
		}

		c.step = RequestedDescription
		sendEscMsg(c.bot, chatID, "Nice one. Now send me the description, please?")

	// step 3: description
	case RequestedDescription:
		if err := c.db.UpdateShortUrl(c.halfSavedShortID, "Description", message.Text); err != nil {
			sendEscMsg(c.bot, chatID, "Sorry, can't update Description "+
				"Can you send me once again please?")
			log.Println("Failed to update a short link, error " + err.Error())
			return
		}

		c.step = RequestedButtonIsPrivateOrPublic
		resp, _ := sendEscMsg(c.bot, chatID, "Nice one. Is it public or private?")
		renderPublicPrivateButtons(c.bot, chatID, resp.MessageID)
	}
}

func (c *Command) ProcessButtonCallback(callbackQuery *tgbotapi.CallbackQuery) {

	// notify the telegram that we processed the button, it will turn "loading indicator" off
	defer c.bot.AnswerCallbackQuery(tgbotapi.CallbackConfig{
		CallbackQueryID: callbackQuery.ID,
	})

	// if that button was "Public" or "Private"
	if c.step == RequestedButtonIsPrivateOrPublic {

		if err := c.db.UpdateShortUrl(c.halfSavedShortID, "IsPublic", callbackQuery.Data == public); err != nil {
			sendEscMsg(c.bot, callbackQuery.Message.Chat.ID, "Sorry, can't update short link "+
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

	// expected data is "command # date", for example
	parts := strings.Split(callbackQuery.Data, Separator)
	if parts[0] == ButtonDeleteMsgPrefix {

		// delete message
		deleteMessage(c.bot, callbackQuery.Message.Chat.ID, parts[1])
	} else if parts[0] == ButtonUpdateMsgPrefix {

		// update message
		c.renderStats(parts[2], callbackQuery.Message.Chat.ID, parts[1])
	}
}

// parses command, like stats5x20201201x6; please refer to unit tests for examples
func extractIdAndDate(command string) (int, time.Time, error) {
	arrParts := patternCommandStatsForURLOneDay.FindStringSubmatch(command)
	shortUrlID, err := strconv.Atoi(arrParts[1])
	if err != nil {
		log.Printf("Cant extract ID from the %s command, error is %s", command, err.Error())
		return 0, time.Now(), err
	}

	dayDate, err := time.Parse("20060102", arrParts[2])
	if err != nil {
		log.Printf("Cant parse %s date from the %s command, error is %s", arrParts[2], command, err.Error())
		return 0, time.Now(), err
	}

	return shortUrlID, dayDate, nil
}

func (c *Command) getStatisticOneView(chatID int64, command string) {

	// extract ID
	arrParts := patternCommandStatsView.FindStringSubmatch(command)
	viewID, err := strconv.Atoi(arrParts[2])
	if err != nil {
		log.Printf("Cant extract ID from the %s command, error is %s", command, err.Error())
		sendMsg(c.bot, chatID, "Cant parse command")
		return
	}

	view, err := c.db.GetViewByID(viewID)
	if err != nil {
		log.Printf("View with ID=%d not found, error is %s", viewID, err.Error())
		sendMsg(c.bot, chatID, "Cant find this view")
		return
	}

	output, ok := renderTemplate(c.bot, chatID, "view.md", view)
	if !ok {
		return
	}

	resp, _ := sendMsg(c.bot, chatID, output)

	dayDate, _ := time.Parse(db.DayFormat, view.Day)
	renderCloseUpdateButton(c.bot, chatID, resp.MessageID, command, dayDate)
}

func (c *Command) getStatisticOneURLOneDay(chatID int64, command string, messageIDtoReplace int) {

	shortUrlID, dayDate, err := extractIdAndDate(command)
	if err != nil {
		sendMsg(c.bot, chatID, "Cant parse date")
		return
	}

	shortURL, statsDay, err := c.db.GetStatisticForOneURLOneDay(shortUrlID, dayDate)
	if err != nil {
		log.Printf("Cant get stats for the short ID=%d from the %s command, error is %s", shortUrlID, command, err.Error())
		sendMsg(c.bot, chatID, "Cant get stats")
		return
	}

	statsData := StatsForURLOneDay{
		Views:        statsDay,
		ShortURL:     *shortURL,
		SelectedDate: dayDate.Format(db.DayFormat),
	}

	output, ok := renderTemplate(c.bot, chatID, "stats.one.day.md", statsData)
	if !ok {
		return
	}

	var resp tgbotapi.Message
	if messageIDtoReplace == 0 {
		resp, _ = sendMsg(c.bot, chatID, output)
	} else {
		resp, _ = updateMsg(c.bot, chatID, messageIDtoReplace, output)
	}

	renderCloseUpdateButton(c.bot, chatID, resp.MessageID, command, dayDate)
}

func (c *Command) getStatisticOneURL(chatID int64, command string, messageIDtoReplace int) {

	// extract ID
	arrParts := patternCommandStatsForURL.FindStringSubmatch(command)
	shortUrlID, err := strconv.Atoi(arrParts[2])
	if err != nil {
		log.Printf("Cant extract ID from the %s command, error is %s", command, err.Error())
		sendMsg(c.bot, chatID, "Cant parse command")
		return
	}

	// find statistics
	sURL, views, err := c.db.GetStatisticsForOneURL(shortUrlID)
	if err != nil {
		log.Printf("Cant get statistics for %s command (short ID = %d), error is %s", command, shortUrlID, err.Error())
		sendMsg(c.bot, chatID, "Cant get statistics")
		return
	}

	statsData := StatsForOneURL{
		Stats:    views,
		ShortURL: *sURL,
	}
	output, ok := renderTemplate(c.bot, chatID, "stats.one.url.md", statsData)
	if !ok {
		return
	}

	if messageIDtoReplace == 0 {
		sendMsg(c.bot, chatID, output)
	} else {
		updateMsg(c.bot, chatID, messageIDtoReplace, output)
	}
}

func (c *Command) printStatisticSummary(chatID int64, messageIDtoReplace int) {

	stats, err := c.db.GetAllStatisticsGroupedByURLs()
	if err != nil {
		log.Printf("Error getting grouped stats, err is %s", err.Error())
		sendMsg(c.bot, chatID, "Error getting grouped stats")
		return
	}

	statsData := StatsGroupedByURLData{
		Stats:    stats,
		Hostname: c.hostname,
	}
	output, ok := renderTemplate(c.bot, chatID, "stats.md", statsData)
	if !ok {
		return
	}

	if messageIDtoReplace == 0 {
		sendMsg(c.bot, chatID, output)
		return
	} else {
		updateMsg(c.bot, chatID, messageIDtoReplace, output)
	}
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
	for _, k := range shortenedUrls {

		// log.Println(k)

		sb.WriteString(strconv.Itoa(k.ID))
		sb.WriteString("\\) [")
		sb.WriteString(markdownEscape(k.ShortUrl))
		sb.WriteString("](")
		sb.WriteString(markdownEscapeUrl(k.TargetUrl))
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

func sendEscMsg(bot *tgbotapi.BotAPI, chatID int64, textMarkdown string) (tgbotapi.Message, error) {
	return sendMsg(bot, chatID, markdownEscape(textMarkdown))
}

// simply send a message to bot in Markdown format
func sendMsg(bot *tgbotapi.BotAPI, chatID int64, textMarkdown string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, textMarkdown)
	msg.ParseMode = "MarkdownV2"
	msg.DisableWebPagePreview = true

	// send the message
	resp, err := bot.Send(msg)
	if err != nil {
		log.Println("bot.Send:", err, resp, textMarkdown)
		return resp, err
	}

	return resp, err
}

// simply send a message to bot in Markdown format
func updateMsg(bot *tgbotapi.BotAPI, chatID int64, messageIDtoReplace int, textMarkdown string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewEditMessageText(chatID, messageIDtoReplace, textMarkdown)
	msg.ParseMode = "MarkdownV2"
	msg.DisableWebPagePreview = true

	resp, err := bot.Send(msg)
	if err != nil {
		log.Println("bot.Send:", err, resp, textMarkdown)
		return resp, err
	}

	return resp, err
}

func renderTemplate(bot *tgbotapi.BotAPI, chatID int64, templateFileName string, data interface{}) (string, bool) {
	var sb strings.Builder
	tplStatsOneDay := template.Must(template.New(templateFileName).Funcs(funcMap).ParseFiles("templates/" + templateFileName))
	if err := tplStatsOneDay.ExecuteTemplate(&sb, templateFileName, data); err != nil {
		log.Printf("error executing template %s, err is %s", templateFileName, err.Error())
		sendMsg(bot, chatID, "Error parsing template")
		return "", false
	}
	return sb.String(), true
}

func renderCloseUpdateButton(bot *tgbotapi.BotAPI, chatID int64, messageID int, command string, dateMidnight time.Time) {
	strMessageID := strconv.Itoa(messageID)

	rowCloseButton := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("❌ Close", ButtonDeleteMsgPrefix+Separator+strMessageID),
	}

	if time.Since(dateMidnight).Hours() < 24 {
		rowCloseButton = append(rowCloseButton, tgbotapi.NewInlineKeyboardButtonData("🔄 Update", ButtonUpdateMsgPrefix+Separator+strMessageID+Separator+command))
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rowCloseButton)
	keyboardMsg := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, keyboard)
	bot.Send(keyboardMsg)
}

func deleteMessage(bot *tgbotapi.BotAPI, chatID int64, messageID string) {
	if intMessageID, err := strconv.Atoi(messageID); err == nil {
		msg := tgbotapi.NewDeleteMessage(chatID, intMessageID)
		bot.Send(msg)
	} else {
		log.Println(err)
	}
}

// From DOCs: https://core.telegram.org/bots/api#markdownv2-style
// esc
// '_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!'
// must be escaped with the preceding character '\'.`'
func markdownEscape(s string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(s)
}

// Inside (...) part of inline link definition,
// all ')' and '\' must be escaped with a preceding '\' character.
func markdownEscapeUrl(s string) string {
	replacer := strings.NewReplacer(
		")", "\\)",
		"\\", "\\\\",
	)
	return replacer.Replace(s)
}
