package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/pkg/errors"
	"github.com/vmarkovtsev/go-lcss"
)

// VoiceConnection contains the actual session and some info about the voice channel
type VoiceConnection struct {
	s       *discordgo.Session
	guild   string
	channel string
}

// JoinAndPlay connects to a voice channel and plays a sound
func JoinAndPlay(s *discordgo.Session, guildID, channelID, name string) error {

	conn := VoiceConnection{
		s:       s,
		guild:   guildID,
		channel: channelID,
	}

	switch name {
	case "wow":
		return nil
	default:
		err := playLocal(conn, name)
		if err != nil {
			return err
		}
	}
	return nil
}

// playWow plays searches the WoWHead DB and plays the cloest sound
func playWow(conn VoiceConnection, name string) error {
	return nil
}

// playLocal plays a local sound file from the `sounds` directory
func playLocal(conn VoiceConnection, name string) error {
	// Only read the first thing as an argument
	f := strings.SplitN(name, " ", 2)

	sound, err := loadSound(f[0])
	if err != nil {
		return err
	}

	// Look for the message sender in that guild's current voice states.
	err = playSound(conn, sound)
	if err != nil {
		return errors.Wrap(err, "error playing sound")
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
func playSound(conn VoiceConnection, sound string) error {

	// Join the provided voice channel.
	vc, err := conn.s.ChannelVoiceJoin(conn.guild, conn.channel, false, true)
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
