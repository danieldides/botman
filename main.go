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

	// User join handler
	// discord.AddHandler(onJoin)

	// Loop to keep the bot alive
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	discord.Close()
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
	args = parts[2]

	switch cmd {
	case "poll":
		log.Printf("Request to create poll received by %s\n", m.Author.Username)
		err := CreatePoll(s, m, args)
		if err != nil {
			log.Printf("Error creating poll: %v\n", err)
		}
	case "are":
		s.ChannelMessageSend(m.ChannelID, "Nope.")
	case "horn":
		err := JoinAndPlay(s, m, args)
		if err != nil {
			log.Println(err)
		}
	default:
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Invalid command: %s", cmd))
	}
}
