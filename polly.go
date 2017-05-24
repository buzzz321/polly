package main

import (
	"bufio"
	"fmt"
	"gw2util"
	"log"
	"os"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	BotID    string
	userData gw2util.UserDataSlice
	dg       *discordgo.Session
	guilds   map[string]*discordgo.Guild // not thread safe..
	mutex    = &sync.Mutex{}
)

func readkey(filename string) string {
	inputFile, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer inputFile.Close()

	scanner := bufio.NewScanner(inputFile)
	scanner.Scan()
	key := scanner.Text()
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	return key
}

// This function will be called (due to AddHandler below) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == BotID {
		return
	}

	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Ping!")
	}

}

func guildCreate(s *discordgo.Session, mguild *discordgo.GuildCreate) {
	mutex.Lock()
	fmt.Printf("Name : %s Id: %s\n", mguild.Name, mguild.ID)
	guilds[mguild.Channels[0].ID] = mguild.Guild

	if mguild.Name == "Notte_test" {
		if dg != nil {
			//dg.ChannelMessageSend(guilds["256795736677679104"].Channels[0].ID, "Tjo!")
			//dg.ChannelMessageSend(guilds["95498187816570880"].Channels[0].ID, "Tjo!")
		}
	}
	mutex.Unlock()
}

func runner() {
	var redKD, greenKD, kDBlue float64 = 0.0, 0.0, 0.0
	gw2 := gw2util.Gw2Api{BaseUrl: "https://api.guildwars2.com/v2/", Key: gw2util.GetUserData(userData, "Notimik").Key}
	for {

		stats := gw2util.GetWWWStats(gw2, "2007")

		if stats.Kills.Blue > 0 {
			kDBlue = stats.Kills.Blue / stats.Deaths.Blue
		}
		if stats.Kills.Red > 0 {
			redKD = stats.Kills.Red / stats.Deaths.Red
		}
		if stats.Kills.Green > 0 {
			greenKD = stats.Kills.Green / stats.Deaths.Green
		}

		mutex.Lock()
		msg := fmt.Sprintf("K/D \n Blue: %6.2f\n Red: %6.2f\n Green: %6.2f\n", kDBlue, redKD, greenKD)
		//fmt.Println(msg)
		if len(guilds) > 0 {
			//dg.ChannelMessageSend(guilds["256795736677679104"].Channels[0].ID, msg)
			dg.ChannelMessageSend(guilds["95498187816570880"].Channels[0].ID, msg)
		}
		mutex.Unlock()
		time.Sleep(20 * time.Minute)
	}
	fmt.Println("End of runner")
}

func main() {
	discordKey := readkey("../../../discord/polly.key")
	guilds = make(map[string]*discordgo.Guild)

	// Create a new Discord session using the provided bot token.
	var err error
	dg, err = discordgo.New("Bot " + discordKey)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	userData = gw2util.ReadUserData("data.dat")
	// Get the account information.
	u, err := dg.User("@me")
	if err != nil {
		fmt.Println("error obtaining account details,", err)
	}
	// Store the account ID for later use.
	BotID = u.ID

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)
	dg.AddHandler(guildCreate)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	//dg.ChannelMessageSend(guilds[1].Channels, "hejsan hoppsan")
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	go runner()
	// Simple way to keep program running until CTRL-C is pressed.
	<-make(chan struct{})

	fmt.Println("Exiting...")
	return
}
