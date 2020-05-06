package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/bwmarrin/discordgo"
)

func main() {
	godotenv.Load()

	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
	if err != nil {
		log.Fatal("unable to create discord client")
	}

	discord.AddHandler(read)

	err = discord.Open()
	if err != nil {
		log.Fatal("unable to open connection to discord")
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	discord.Close()
}

func read(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	fmt.Println(m.Content)

	if m.Content == "!stalk-prices" {
		s.ChannelMessageSend(m.ChannelID, "Island Code: DOG818 Price: 489 Bells")
	}

	if m.Content == "!stalk-register" {
		s.ChannelMessageSend(m.ChannelID, "Oops! Make sure you include your island code and price - (!stalk-register DOG827 429)")
	}

	if m.Content == "!stalk-help" {
		s.ChannelMessageSend(m.ChannelID,
			`Stalk-bot allows you to register your island's'`)
	}
}
