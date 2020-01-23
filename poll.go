package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	// PollReactions is a list of emoji reactions for poll feature
	PollReactions = [...]string{
		"🚀", "🍍", "🐶", "💀",
		"🍺", "🚒", "🐱", "🐉",
		"👿", "🏫", "👾", "👹",
		"🍔", "🍟", "🍣", "🍤",
		"🍇", "🍅", "🍑", "🍌",
		"🍍", "🥓", "🥔", "🥕",
	}
)

// USAGE is the basic invocation string for starting a poll
const USAGE = "Usage: `@botman [title] [duration s] [option1,option2,...]`"

// Poll is a new Poll
type Poll struct {
	title    string
	duration time.Duration
	choices  []string
	err      error
}

func parseOpts(body string) Poll {
	var p Poll

	r := csv.NewReader(strings.NewReader(body))
	r.Comma = ' '
	fields, err := r.Read()
	if err != nil {
		p.err = err
		return p
	}

	if len(fields) < 3 {
		p.err = errors.New("not enough arguments")
		return p
	}

	p.title = fields[0]

	s, err := strconv.Atoi(fields[1])
	if err != nil {
		p.err = errors.New("time must be specified in seconds")
		return p
	}
	p.duration = time.Duration(s) * time.Second
	p.choices = fields[2:]

	return p
}

// CreatePoll creates a poll of `opts...` with voting by reactions. After the duration, it returns the winner
func CreatePoll(s *discordgo.Session, m *discordgo.MessageCreate, body string) error {
	p := parseOpts(body)

	if p.err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: %s.\n %s\n", p.err, USAGE))
		return p.err
	}

	message := fmt.Sprintf("%s has begun a poll: **%s** \nVote below using the reactions.\n", m.Author.Mention(), p.title)
	for idx, choice := range p.choices {
		message += fmt.Sprintf("%v:\t%s\n", PollReactions[idx], choice)
	}
	pm, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		return err
	}
	for i := 0; i < len(p.choices) && i < 20; i++ {
		s.MessageReactionAdd(m.ChannelID, pm.ID, PollReactions[i])
	}

	if p.duration != 0 {
		time.AfterFunc(p.duration, func() {
			var winner int

			poll, _ := s.ChannelMessage(pm.ChannelID, pm.ID)

			// Calculate which reaction has the most votes
			max := 0
			for i, reaction := range poll.Reactions {
				if reaction.Count > max {
					max = reaction.Count
					winner = i
				}
			}
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Voting closed. Result is:\n \t\t %s: %s\n", PollReactions[winner], p.choices[winner]))
		})
	}

	return nil
}
