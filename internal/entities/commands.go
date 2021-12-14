package entities

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

// Command is a value for UserCommands map.
type Command struct {
	Id          uint64
	Description string
	Command     string
	Pattern     string
	MinRange    int
	MaxRange    int
}

// UserCommand is commands from user.
type UserCommand struct {
	Ttl    time.Time
	Chat   *tgbotapi.Chat
	Cmd    Command
	Result string
}

// UserCommands is key are UserId of Users
type UserCommands map[int64][]map[string]UserCommand

// Commands is array of commands where key is a command
type Commands map[string]Command
