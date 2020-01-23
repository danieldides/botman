package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// JoinAndPlay connects to a voice channel and plays a sound
func JoinAndPlay(s *discordgo.Session, m *discordgo.MessageCreate, name string) error {

	// Only read the first thing as an argument
	f := strings.SplitN(name, " ", 2)
	buffer, err := loadSound(f[0])
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
			err = playSound(s, g.ID, vs.ChannelID, buffer)
			if err != nil {
				fmt.Println("Error playing sound:", err)
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
func loadSound(name string) ([][]byte, error) {
	var buffer = make([][]byte, 0)
	var file string

	const path = "sounds"

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	type kv struct {
		Key   string
		Value int
	}

	// Iterate through all the sounds in sounds/ and find the one most
	// like the `name` argument
	var similarities map[string]int

	for _, f := range files {
		base := filepath.Base(f.Name())
		_ = base
		similarities[f.Name()] = 0
	}

	var ss []kv
	for k, v := range similarities {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	file, err := os.Open(ss[0].Key)
	if err != nil {
		fmt.Println("Error opening dca file :", err)
		return nil, err
	}

	var opuslen int16

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := file.Close()
			if err != nil {
				return nil, err
			}
			return nil, nil
		}

		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return nil, err
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return nil, err
		}

		// Append encoded pcm data to the buffer.
		buffer = append(buffer, InBuf)
	}
	return buffer, nil
}

// playSound plays the current buffer to the provided channel.
func playSound(s *discordgo.Session, guildID, channelID string, buffer [][]byte) error {

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.
	vc.Speaking(true)

	// Send the buffer data.
	for _, buff := range buffer {
		vc.OpusSend <- buff
	}

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	// Disconnect from the provided voice channel.
	vc.Disconnect()

	return nil
}
