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
	"sizebot/internal/config"
	"sizebot/internal/entities"
	"sizebot/internal/storage/postgres"
	"strconv"
	"strings"
	"time"
)

//const ttlCommandForUser = time.Hour * 24
const ttlCommandForUser = time.Minute

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln("No .env file found")
	}
}

func main() {

	appConfig := config.New()
	ctx := grace.ShutdownContext(context.Background())

	pg, err := postgres.New(ctx, &appConfig.Postgres)
	if err != nil {
		log.Fatalln("failed to init postgres")
	}

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
	userCommand := make(map[string]entities.UserCommand)

	for update := range updates {
		if update.Message != nil { // If we got a message
			text := strings.ReplaceAll(update.Message.Text, "@all_size_of_bot", "")
			command, ok := commands[text]
			if !ok {
				continue
			}

			commandExist := false
			for i := range userCommands[update.Message.From.ID] {
				commandUser, exist := userCommands[update.Message.From.ID][i][command.Command]
				if exist {
					timeLeft := int(commandUser.Ttl.Sub(time.Now()) / time.Minute)
					if timeLeft > 0 {
						commandExist = true
						msg := tgbotapi.NewMessage(
							update.Message.Chat.ID,
							commandUser.Result,
						)
						msg.ReplyToMessageID = update.Message.MessageID
						bot.Send(msg)
						break
					}
					break
				}
			}

			if commandExist {
				continue
			}

			r := randomSize(rand.Intn(command.MinRange), rand.Intn(command.MaxRange))
			result := fmt.Sprintf(command.Pattern, fmt.Sprintf("%.2f", r), getEmoji(r, command))
			userCommand[command.Command] = entities.UserCommand{
				Ttl:    time.Now().Add(ttlCommandForUser),
				Chat:   update.Message.Chat,
				Cmd:    command,
				Result: result,
			}
			userCommands[update.Message.From.ID] = append(
				userCommands[update.Message.From.ID],
				userCommand,
			)

			msg := tgbotapi.NewMessage(
				update.Message.Chat.ID,
				result,
			)
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
	}

	pg.Close()
}

func randomSize(a int, b int) float32 {
	return float32(a+b) + rand.Float32()
}

func getEmoji(n float32, c entities.Command) string {
	middle := float32(c.MaxRange / 2)
	oneThird := float32(c.MaxRange / 3)
	doubleOneThird := float32(oneThird * 2)

	emoji := "\\U0001F613"
	if n > oneThird && n <= middle {
		emoji = "\\U0001F623"
	} else if n > middle && n <= doubleOneThird {
		emoji = "\\U0001F63C"
	} else if n > doubleOneThird {
		emoji = "\\U0001F631"
	}

	// Hex String
	h := strings.ReplaceAll(emoji, "\\U", "0x")

	// Hex to Int
	i, _ := strconv.ParseInt(h, 0, 64)

	// Unescape the string (HTML Entity -> String).
	return html.UnescapeString(string(i))
}
