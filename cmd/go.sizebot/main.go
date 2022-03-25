package main

import (
	"context"
	"fmt"
	"github.com/chapsuk/grace"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"html"
	"log"
	"math/rand"
	"regexp"
	"sizebot/internal/config"
	"sizebot/internal/entities"
	"sizebot/internal/storage/postgres"
	"strconv"
	"strings"
	"time"
)

const ttlCommandForUser = time.Minute * 30

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln("No .env file found")
	}
}
func main() {
	rand.Seed(time.Now().UnixNano())

	appConfig := config.New()
	ctx := grace.ShutdownContext(context.Background())

	pg, err := postgres.New(ctx, &appConfig.Postgres)
	if err != nil {
		log.Fatalln("failed to init postgres")
	}
	defer pg.Close()

	commands, err := pg.Commands(ctx)
	if err != nil {
		log.Fatalln("failed to get commands")
	}

	bot, err := tgbotapi.NewBotAPI(appConfig.Telegram.Token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = appConfig.Telegram.BotDebug
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	userCommands := make(entities.UserCommands)
	inlineQueries := make(chan tgbotapi.InlineQuery, 10)

	go func(inlineQueries chan tgbotapi.InlineQuery, commands entities.Commands) {
		for inlineQuery := range inlineQueries {
			params := getParamsForInlineConfig(
				inlineQuery.From.ID,
				inlineQuery.Query,
				commands,
				userCommands,
			)

			inlineConf := tgbotapi.InlineConfig{
				InlineQueryID: inlineQuery.ID,
				Results:       params,
				CacheTime:     1,
			}
			bot.Send(inlineConf)
		}
	}(inlineQueries, commands)

	go func(updates tgbotapi.UpdatesChannel, commands entities.Commands) {
		for update := range updates {
			if update.Message != nil {
				msg := updateMessageForUser(update, userCommands, commands)
				bot.Send(msg)
				continue
			}
			if update.InlineQuery != nil {
				inlineQueries <- *update.InlineQuery
				continue
			}
		}
	}(updates, commands)
}

// Create inline config.
func getParamsForInlineConfig(
	id int64,
	query string,
	commands entities.Commands,
	userCommands entities.UserCommands,
) []interface{} {
	var params []interface{}
	for key, command := range commands {
		if !strings.Contains(key, query) && !strings.Contains(command.Description, query) {
			continue
		}

		commandUser := getUserCommand(id, &command, userCommands)
		params = append(
			params,
			tgbotapi.NewInlineQueryResultArticle(
				strconv.FormatUint(command.Id, 10),
				command.Description,
				commandUser.Result,
			),
		)
	}

	return params
}

func updateMessageForUser(update tgbotapi.Update, userCommands entities.UserCommands, commands entities.Commands) *tgbotapi.MessageConfig {
	re := regexp.MustCompile(`@.*`)
	text := re.ReplaceAllString(update.Message.Text, "")
	command, ok := commands[text]
	if !ok {
		return nil
	}

	commandUser := getUserCommand(update.Message.From.ID, &command, userCommands)
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		commandUser.Result,
	)
	msg.ReplyToMessageID = update.Message.MessageID

	return &msg
}

func getUserCommand(fromId int64, command *entities.Command, userCommands entities.UserCommands) *entities.UserCommand {
	commandUser := getCommandUserById(fromId, command.Command, userCommands)
	if commandUser == nil {
		commandUser = addCommandInUserFromId(fromId, command, userCommands)
		return commandUser
	}

	timeLeft := int(commandUser.Ttl.Sub(time.Now()) / time.Minute)
	if timeLeft <= 0 {
		commandUser = addCommandInUserFromId(fromId, command, userCommands)
	}

	return commandUser
}
func getCommandUserById(id int64, command string, userCommands entities.UserCommands) *entities.UserCommand {
	for i := range userCommands[id] {
		commandUser, ok := userCommands[id][i][command]
		if !ok {
			continue
		}

		return &commandUser
	}

	return nil
}
func addCommandInUserFromId(id int64, command *entities.Command, userCommands entities.UserCommands) *entities.UserCommand {

	r := randomSize(rand.Intn(command.MinRange), rand.Intn(command.MaxRange))
	result := fmt.Sprintf(command.Pattern, fmt.Sprintf("%.2f", r), getEmoji(r, command))

	userCommandMap := make(map[string]entities.UserCommand)
	userCommand := entities.UserCommand{
		Ttl:    time.Now().Add(ttlCommandForUser),
		Cmd:    *command,
		Result: result,
	}
	userCommandMap[command.Command] = userCommand
	userCommands[id] = append(
		userCommands[id],
		userCommandMap,
	)

	return &userCommand
}

func randomSize(a int, b int) float32 {
	return float32(a+b) + rand.Float32()
}
func getEmoji(n float32, c *entities.Command) string {
	middle := float32(c.MaxRange / 2)
	oneThird := float32(c.MaxRange / 3)
	doubleOneThird := oneThird * 2
	var emoji string

	switch {
	case n >= doubleOneThird:
		emoji = "\\U0001F631"
	case n >= middle:
		emoji = "\\U0001F63C"
	case n >= oneThird:
		emoji = "\\U0001F63C"
	default:
		emoji = "\\U0001F613"
	}

	// Hex String.
	h := strings.ReplaceAll(emoji, "\\U", "0x")

	// Hex to Int
	i, _ := strconv.ParseInt(h, 0, 64)

	// Unescape the string (HTML Entity -> String).
	return html.UnescapeString(string(i))
}
