package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func main() {
	token := os.Getenv("TOKEN")
	bot, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Unable to init bot: %v", err)
	}

	bot.AddHandler(messageCreate)
	bot.AddHandler(memberList)

	bot.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates

	err = bot.Open()
	defer bot.Close()
	if err != nil {
		log.Fatalln("unable to open ws connection, ", err)
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	<-sc
}

func memberList(s *discordgo.Session, c *discordgo.GuildMembersChunk) {
	log.Println("Member chuck")
	for _, member := range c.Members {
		log.Println("Member %+v", member)
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Don't do anything with our own messages
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!join" {

		err := s.RequestGuildMembers(m.GuildID, "", 0, false)
		if err != nil {
			log.Println("Unable to request members")
		}
		channelList, err := s.GuildChannels(m.GuildID)
		if err != nil {
			log.Println("Unable to get channels")
		}
		chanNames := make([]string, 0, len(channelList))
		var funhouse *discordgo.Channel
		for _, channel := range channelList {
			if channel.Type == discordgo.ChannelTypeGuildVoice {
				if channel.Name == "HIJ's Funhouse" {
					funhouse = channel
				}
				chanNames = append(chanNames, channel.Name)
			}
		}
		_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Found these voice channels %v", chanNames))
		if err != nil {
			log.Println("Unable to send message")
		}
		if funhouse == nil {
			log.Println("Unable to find funhouse")
		}
		voice, err := s.ChannelVoiceJoin(funhouse.GuildID, funhouse.ID, false, false)
		if voice.Ready {
			log.Println("I COULD TALK")
		}
		go func() {
			for {
				select {
				case packet := <-voice.OpusRecv:
					log.Println(packet)
				}
			}
		}()
	}
}
