package discord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

type DiscordBot struct {
	session  *discordgo.Session
	log      *logrus.Entry
	commands map[string]CommandCallback
}
type CommandCallback func(*DiscordBot, []string, string, string)

func NewDiscordBot(token string, log *logrus.Entry) (*DiscordBot, error) {
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	bot := &DiscordBot{
		session:  discord,
		log:      log.WithField("module", "discord"),
		commands: make(map[string]CommandCallback),
	}

	discord.AddHandler(bot.messageCreate)

	//identify what we intend to
	discord.Identify.Intents = discordgo.IntentsGuildMessages

	err = discord.Open()
	if err != nil {
		return nil, err
	}

	bot.AddCommand("ping", pingCommand)

	return bot, nil
}

//events
func (b *DiscordBot) onReady(s *discordgo.Session, r *discordgo.Ready) {
	b.log.Debugln("Discord bot ready!")
}
func (b *DiscordBot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return // bot created this message
	}

	if len(m.Content) > 0 {
		if m.Content[0:1] == "!" {
			parts := strings.Split(m.Content[0:], " ")
			command := strings.ToLower(parts[0])

			if b.commands[command] != nil {

				args := make([]string, 0)
				if len(parts) > 1 {
					args = parts[1:]
				}

				go b.commands[command](b, args, m.Author.ID, m.ChannelID)

			}
		}
	}
}

// public methods
func (b *DiscordBot) Shutdown() error {
	return b.session.Close()
}

func (b *DiscordBot) AddCommand(command string, callback CommandCallback) {
	b.commands[command] = callback
}
func (b *DiscordBot) MessageChannel(channelid string, message string) error {
	_, err := b.session.ChannelMessageSend(channelid, message)
	return err
}
func (b *DiscordBot) MessageUser(userid string, message string) error {
	ch, err := b.session.UserChannelCreate(userid)
	if err != nil {
		return err
	}
	_, err = b.session.ChannelMessageSend(ch.ID, message)
	return err
}

// internal methods
func pingCommand(bot *DiscordBot, args []string, sender_id string, channel_id string) {
	bot.MessageChannel(channel_id, "Pong!")
}
