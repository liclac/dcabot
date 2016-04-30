package main

import (
	"encoding/binary"
	"github.com/bwmarrin/discordgo"
	"github.com/codegangsta/cli"
	"io"
	"log"
	"os"
	"os/signal"
)

func action(c *cli.Context) {
	file := c.String("file")
	token := c.String("token")

	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("Couldn't open file", err)
	}

	sess, err := discordgo.New(token)
	if err != nil {
		log.Fatalf("Couldn't create session: %s\n", err)
	}

	sess.AddHandler(func(s *discordgo.Session, e *discordgo.Ready) {
		log.Printf("Ready\n")
	})
	sess.AddHandler(func(s *discordgo.Session, e *discordgo.Disconnect) {
		log.Fatalf("Disconnected\n")
	})
	sess.AddHandler(func(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
		log.Printf("User: %s, Channel: %s\n", vs.UserID, vs.ChannelID)

		vc, err := s.ChannelVoiceJoin(vs.GuildID, vs.ChannelID, false, false)
		if err != nil {
			log.Fatalf("Couldn't join voice: %s\n", err)
		}

		vc.Speaking(true)
		defer vc.Speaking(false)

		for {
			size := int16(0)
			if err := binary.Read(f, binary.LittleEndian, &size); err != nil {
				if err != io.EOF {
					log.Fatalf("Couldn't read length", err)
				}
				break
			}

			data := make([]byte, size)
			if err := binary.Read(f, binary.LittleEndian, &data); err != nil {
				log.Fatalf("Couldn't read data", err)
			}
			vc.OpusSend <- data
		}

		log.Printf("Done!\n")
		os.Exit(0)
	})

	if err = sess.Open(); err != nil {
		log.Fatalf("Couldn't open connection: %s\n", err)
		return
	}

	exit := make(chan os.Signal)
	signal.Notify(exit, os.Interrupt)
	<-exit
	print("\n")
}

// This is hacked together and terrible, don't judge me
func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "file, f"},
		cli.StringFlag{Name: "token, t"},
	}
	app.Action = action
	app.Run(os.Args)
}
