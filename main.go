package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/bwmarrin/discordgo"
)

var args struct {
	Token string
	Url   string `default:"https://moviepolls.zorchenhimer.com/"`
}

func main() {
	arg.MustParse(&args)

	discord, err := discordgo.New("Bot " + args.Token)
	if err != nil {
		log.Fatalf("Could not start discord bot: %v\n", err)
	}

	discord.AddHandler(messageCreate)

	discord.Identify.Intents = discordgo.IntentsGuildMessages

	err = discord.Open()
	if err != nil {
		log.Fatalf("Could not open bot connection: %v\n", err)
	}

	fmt.Println("Bot is running")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-ch

	discord.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	log.Println(m.Content)

	if m.Content == "!lastmovie" {
		date, err := getLastMovieNight()
		if err != nil {
			fmt.Printf("Error: could not get time since last movienight: %v\n", err)
			return
		}

		_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("It has been <t:%d:R> days since MovieNight was last done on <t:%d:D>", date.Unix(), date.Unix()))
		if err != nil {
			fmt.Printf("ERROR: issue sending message: %v\n", err)
			return
		}

		fmt.Printf("Message requested by %#v sent\n", m.Author.Username)
	}
}

func getLastMovieNight() (time.Time, error) {
	resp, err := http.Get(args.Url)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not get url: %w", err)
	}
	defer resp.Body.Close()

	htmlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not read body: %w", err)
	}
	html := string(htmlBytes[:])

	re, _ := regexp.Compile("In the last Movienight on (.+) we watched:")
	dateString := re.FindStringSubmatch(html)[1]

	date, err := time.Parse("Mon Jan 02, 2006", dateString)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not parse date: %w", err)
	}

	return date, nil
}
