package main

import (
	"fmt"
	"log"
	"mcCoordsBot/locations"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

const (
	FileName    string = "locations.json"
	Version     string = "V1.1"
	CmdPrefix   string = "#"
	CmdSave     string = "save"
	CmdDelete   string = "delete"
	CmdList     string = "list"
	CmdVersion  string = "version"
	CmdHelp     string = "help"
	CmdHelpBody string = "```" + "#save <name> <x> <y> <z>: save a location\n" +
		"#delete <name>: delete a location\n" +
		"#list: Show all saved locations\n" +
		"#help: Display help\n" +
		"#version: Display bot version\n" +
		"```"
)

var (
	Token string
	lmap  locations.LocationMap
)

//Calls a function every 'delay' time units
func repeat(fn func(), delay time.Duration) chan bool {
	stop := make(chan bool)

	go func() {
		for {
			fn()
			select {
			case <-time.After(delay):
			case <-stop:
				return
			}
		}
	}()

	return stop
}

func init() {
	var err error

	err = godotenv.Load(".env")
	if err != nil {
		log.Printf("Could not load .env file: %s\n", err.Error())
	}

	Token = os.Getenv("DISCORD_TOKEN")

	log.SetFlags(log.Lshortfile)

	lmap, err = locations.Load(FileName)
	if err != nil {
		lmap = locations.New()
		log.Printf("No saved locations found: %s\n", err.Error())
	}
}

func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatalf("Error creating discord session: %s\n", err.Error())
		return
	}

	dg.AddHandler(ready)
	dg.AddHandler(messageCreate)

	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %s", err.Error())
		return
	}

	stop := repeat(func() {
		lmap.Save(FileName)
	}, time.Minute*30)

	log.Printf("Bot is running...\n")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	lmap.Save(FileName)
	stop <- true
	dg.Close()
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateStatus(1, "Covfefe")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID { //Message is from this bot
		return
	}

	if strings.Contains(m.Content, "ET") {
		s.MessageReactionAdd(m.ChannelID, m.ID, "🍗")
	}

	parts := strings.Split(m.Content, " ")
	partCount := len(parts)

	if parts[0] == CmdPrefix+CmdSave {

		if partCount != 5 { //Invalid argument count
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("`Invalid argument count, need 5 but got %d`", partCount))
			return
		}

		name := parts[1]

		x, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "`X coordinate is not a number`")
			return
		}

		y, err := strconv.ParseFloat(parts[3], 64)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "`Y coordinate is not a number`")
			return
		}

		z, err := strconv.ParseFloat(parts[4], 64)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "`Z coordinate is not a number`")
			return
		}

		lmap.Set(name, locations.Location{x, y, z})
		s.ChannelMessageSend(m.ChannelID, "`Saved Location!`")

	} else if parts[0] == CmdPrefix+CmdDelete {
		if partCount != 2 {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("`Invalid argument count, need 2 but got %d`", partCount))
			return
		}

		name := parts[1]
		err := lmap.Delete(name)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("`Couldn't delete because: %s`", err.Error()))
			return
		}
		s.ChannelMessageSend(m.ChannelID, "`Deleted "+name+"!`")

	} else if parts[0] == CmdPrefix+CmdList {
		list := lmap.ToString()
		if len(list) < 1 {
			list = "Use #save to add a location!"
		}
		s.ChannelMessageSend(m.ChannelID, "```"+list+"```")
	} else if parts[0] == CmdPrefix+CmdVersion {
		s.ChannelMessageSend(m.ChannelID, "`"+Version+"`")
	} else if parts[0] == CmdPrefix+CmdHelp {
		s.ChannelMessageSend(m.ChannelID, CmdHelpBody)
	}
}
