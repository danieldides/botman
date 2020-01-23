package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/vmarkovtsev/go-lcss"
)

// JoinAndPlay connects to a voice channel and plays a sound
func JoinAndPlay(s *discordgo.Session, m *discordgo.MessageCreate, name string) error {

	// Only read the first thing as an argument
	f := strings.SplitN(name, " ", 2)

	sound, err := loadSound(f[0])
	if err != nil {
		return err
	}

	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		log.Println(err)
	}

	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		log.Println(err)
	}

	// Look for the message sender in that guild's current voice states.
	for _, vs := range g.VoiceStates {
		if vs.UserID == m.Author.ID {
			err = playSound(s, g.ID, vs.ChannelID, sound)
			if err != nil {
				return fmt.Errorf("Error playing sound %s: %v", sound, err)
			}
			return nil
		}
	}
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

// loadSound attempts to load an encoded sound file from disk.
func loadSound(name string) (string, error) {
	const dir = "sounds"

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}

	type kv struct {
		Key   string
		Value int
	}

	// Iterate through all the sounds in sounds/ and find the one most
	// like the `name` argument
	similarities := make(map[string]int)

	for _, f := range files {
		base := filepath.Base(f.Name())
		similarities[f.Name()] = len(lcss.LongestCommonSubstring([]byte(base), []byte(name)))
	}

	var ss []kv
	for k, v := range similarities {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	return filepath.Join(dir, ss[0].Key), nil
}

// playSound plays the current buffer to the provided channel.
func playSound(s *discordgo.Session, guildID, channelID, sound string) error {

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.
	vc.Speaking(true)

	// if filepath.Ext(sound) != "dca" {
	encoder, err := dca.EncodeFile(sound, dca.StdEncodeOptions)
	if err != nil {
		return fmt.Errorf("Error encoding file %s to DCA: %v", sound, err)
	}
	defer encoder.Cleanup()

	done := make(chan error)

	dca.NewStream(encoder, vc, done)
	err = <-done
	if err != nil && err != io.EOF {
		return err
	}
	// }

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	// Disconnect from the provided voice channel.
	vc.Disconnect()

	return nil
}
