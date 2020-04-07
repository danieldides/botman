package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
)

const COMMANDS = "poll"

var state *discordgo.State

func main() {
	token := os.Getenv("DISCORD_TOKEN")

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}

	err = discord.Open()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Logged in as " + discord.State.User.Username)

	// Poll handler
	discord.AddHandler(onMessage)

	// Voice State Update handler
	// discord.AddHandler(onJoin)

	// Loop to keep the bot alive
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	discord.Close()
}

func onJoin(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
	var mumble = os.Getenv("MUMBLE_ID")

	// Ignore botman joining events
	if vs.UserID == s.State.User.ID {
		return
	}

	// Ignore messages in other guilds
	if vs.GuildID != mumble {
		return
	}

	// Fixed friend IDs
	crychair := os.Getenv("CRYCHAIR_ID")
	senque := os.Getenv("SENQUE_ID")
	steggo := os.Getenv("STEGGO_ID")
	vinny := os.Getenv("VINNY_ID")

	// Don't play if Dan's in the chat
	g, err := s.Guild(vs.GuildID)
	if err != nil {
		log.Println(err)
		return
	}
	for _, voiceState := range g.VoiceStates {
		if voiceState.UserID == steggo {
			log.Println("Dan's here EVERYONE BE QUIET")
			return
		}
	}

	if vs.UserID == crychair || vs.UserID == senque || vs.UserID == vinny {
		err := JoinAndPlay(s, vs.GuildID, vs.ChannelID, "zoop.mp3")
		if err != nil {
			log.Println(err)
			return
		}
	}
	log.Println(vs.UserID)
}

// Parse out messages for commands and call out to functions to handle actions
func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore our own messages
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore all messages that don't ping the bot
	if !(len(m.Mentions) > 0 && m.Mentions[0].ID == s.State.User.ID) {
		return
	}

	// Heart good bots
	if strings.Contains(strings.ToLower(m.Content), "good bot") {
		s.MessageReactionAdd(m.ChannelID, m.ID, "♥")
		return
	}

	// Fuck you bad bots
	if strings.Contains(strings.ToLower(m.Content), "bad bot") {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Fuck you %s.", m.Author.Mention()))
		return
	}

	// bbq
	if strings.Contains(strings.ToLower(m.Content), "bbq sauce sandwich") {
		s.ChannelMessageSend(m.ChannelID, "🤮")
		return
	}

	// Message format: @botman cmd args...
	parts := strings.SplitN(m.Content, " ", 3)

	if len(parts) < 2 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Please provide a command. Commands: %s\n", COMMANDS))
		return
	}

	var args string

	cmd := parts[1]
	if len(parts) < 3 {
		parts = append(parts, "")
	}
	args = parts[2]

	log.Printf("Message command %s(%s) from %s", cmd, args, m.Author.Username)

	switch cmd {
	case "poll":
		log.Printf("Request to create poll received by %s\n", m.Author.Username)
		err := CreatePoll(s, m, args)
		if err != nil {
			return
			// return errors.Wrap(err, "error creating poll")
		}
	case "are":
		s.ChannelMessageSend(m.ChannelID, "Nope.")
	case "play":
		// Determine voice channel that summoner is part of
		c, err := getVoiceChannel(s, m)
		if err != nil {
			log.Printf("error getting voice channel %s\n", err)
		}

		err = JoinAndPlay(s, m.GuildID, c, args)
		if err != nil {
			return
		}

	case "start":
		parts := strings.Split(args, " ")
		if len(parts) > 0 {
			if parts[0] != "video" {
				return
			}
			// Create the URL
			const url = "https://discordapp.com/channels/%s/%s"

			c, err := getVoiceChannel(s, m)
			if err != nil {
				log.Printf("error starting video chat %s\n", err)
			}
			if c == "" {
				s.ChannelMessageSend(m.ChannelID, "You are not in a voice channel. Please join a voice channel and try again.")
				return
			}
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(url, m.GuildID, c))
			return
		}

	default:
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Invalid command: %s", cmd))
	}
	// return nil
	return
}

func getVoiceChannel(s *discordgo.Session, m *discordgo.MessageCreate) (string, error) {
	// Determine server guild
	g, err := s.State.Guild(m.GuildID)
	if err != nil {
		return "", err
	}

	c := ""
	for _, vs := range g.VoiceStates {
		if vs.UserID == m.Author.ID {
			c = vs.ChannelID
		}
	}
	return c, nil
}
