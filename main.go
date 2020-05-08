package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/bwmarrin/discordgo"
)

type IslandPrice struct {
	IslandCode  string `json:"island_code"`
	TurnipPrice int    `json:"turnip_price"`
	Error       string `json:"error,omitempty"`
}

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

	if m.Content == "!stalk-price" {
		islandPrice, err := getTurnipPrices()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error fetching prices %s %v", islandPrice.Error, err))
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Dodo Code: %s Price: %d Bells", islandPrice.IslandCode, islandPrice.TurnipPrice))
		return
	}

	if strings.Contains(m.Content, "!stalk-register") {
		message := strings.Split(m.Content, " ")

		if len(message) < 3 {
			s.ChannelMessageSend(m.ChannelID, "Oops! Make sure you include your dodo code and price - (!stalk-register DOG827 429)")
			return
		}

		price, err := strconv.Atoi(message[2])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Turnip price must be a number")
			return
		}

		err = registerTurnipPrice(message[1], price)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error registering island %s", err.Error()))
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Island successfully registered. Your registration is valid for 30 minutes")
	}

	if m.Content == "!stalk-help" {
		s.ChannelMessageSend(m.ChannelID,
			`Stalk-bot allows you to share your island's turnip prices and invite others to visit at the same time.

Register your island by typing *!stalk-register dodo-code price*.

Ask for an island to visit by typing *!stalk-price*.`)
	}
}

func getTurnipPrices() (IslandPrice, error) {
	islandPrice := IslandPrice{}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", os.Getenv("STALK_API_URL"), nil)
	req.Header.Set("x-api-key", os.Getenv("STALK_API_TOKEN"))

	resp, err := client.Do(req)
	if err != nil {
		return islandPrice, err
	}

	if resp.Body != nil {
		err := json.NewDecoder(resp.Body).Decode(&islandPrice)
		if err != nil {
			return islandPrice, err
		}
	}

	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		return islandPrice, fmt.Errorf("stalk-bot service returned an error")
	}

	if islandPrice.Error != "" {
		return islandPrice, fmt.Errorf("%s", islandPrice.Error)
	}

	return islandPrice, nil
}

func registerTurnipPrice(islandCode string, price int) error {
	islandPrice := IslandPrice{IslandCode: islandCode, TurnipPrice: price}

	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(islandPrice)
	if err != nil {
		return err
	}

	client := &http.Client{}
	req, _ := http.NewRequest("POST", os.Getenv("STALK_API_URL"), buf)
	req.Header.Set("x-api-key", os.Getenv("STALK_API_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		// check for body
		if resp.Body != nil {
			err := json.NewDecoder(resp.Body).Decode(&islandPrice)
			if err == nil {
				return errors.New(islandPrice.Error)
			}
		}

		return errors.New("unable to register island")
	}

	return nil
}
