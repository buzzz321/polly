package main

import (
	"bufio"
	"fmt"
	"gw2util"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	botID    string
	userData gw2util.UserDataSlice
	dg       *discordgo.Session
	guilds   map[string]*discordgo.Guild // not thread safe..
	mutex    = &sync.Mutex{}
)

const notteTestSrv = "256795736677679104"
const sveaUlvarSrv = "95498187816570880"

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
	if m.Author.ID == botID {
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
			//dg.ChannelMessageSend(guilds[notteTestSrv].Channels[0].ID, "Tjo!")
			//dg.ChannelMessageSend(guilds["95498187816570880"].Channels[0].ID, "Tjo!")
		}
	}
	mutex.Unlock()
}

func deleteOldStatsInChannel(chID string) {
	messages, err := dg.ChannelMessages(chID, 100, "", "", "")
	if err != nil {
		fmt.Printf("Couldnt get messages for %s %s\n", chID, err)
	}

	for _, msg := range messages {
		if msg.Author.ID == botID {
			dg.ChannelMessageDelete(chID, msg.ID)
		}
	}
}

func runner() {
	var redKD, greenKD, dBlue float64 = 0.0, 0.0, 0.0
	var wIds string
	var notteMsg, sveaUlvarMsg *discordgo.Message
	var startupStats [5]gw2util.GW2WvWvWStats

	gw2 := gw2util.Gw2Api{BaseURL: "https://api.guildwars2.com/v2/", Key: gw2util.GetUserData(userData, "Notimik").Key}
	serverColours := gw2util.GetWvWvWColours(gw2, "2007")

	for key := range serverColours.WorldColour {
		wIds += strconv.FormatInt(key, 10) + ","
	}
	wIds = wIds[0 : len(wIds)-1]
	serveNames := gw2util.GetWorlds(gw2, wIds)

	var colourName map[string]string
	colourName = make(map[string]string)
	homeWorld := gw2util.GetHomeWorld(gw2)
	fmt.Printf("Home world = %s \n", homeWorld)
	for id, colour := range serverColours.WorldColour {
		colourName[colour] = serveNames.WorldName[id]
	}
	for {

		stats := gw2util.GetWWWStats(gw2, homeWorld /*"2007"*/)
		if startupStats[0].Name == "" {
			fmt.Printf("Saving Old wvwvw stats\n")
			startupStats = stats
		}

		var msg string
		for index, stat := range stats {
			var currRedKD, currGreenKD, currBlueKD float64 = 0.0, 0.0, 0.0
			name := colourName[stat.Name]

			if name == "" {
				name = stat.Name
			}
			if stat.Kills.Blue > 0 {
				dBlue = stat.Kills.Blue / stat.Deaths.Blue
				// add one to denom to combat div by zero ugly but what the heck..
				currBlueKD = (stat.Kills.Blue - startupStats[index].Kills.Blue) / (stat.Deaths.Blue - startupStats[index].Deaths.Blue + 1)
				//fmt.Printf("currk:%6.0f oldk:%6.0f currd:%6.0f oldd:%6.0f\n", stat.Kills.Blue, startupStats[index].Kills.Blue, stat.Deaths.Blue, startupStats[index].Deaths.Blue)
			}
			if stat.Kills.Red > 0 {
				redKD = stat.Kills.Red / stat.Deaths.Red
				currRedKD = (stat.Kills.Red - startupStats[index].Kills.Red) / (stat.Deaths.Red - startupStats[index].Deaths.Red + 1)
				//fmt.Printf("currk:%6.0f oldk:%6.0f currd:%6.0f oldd:%6.0f\n", stat.Kills.Red, startupStats[index].Kills.Red, stat.Deaths.Red, startupStats[index].Deaths.Red)
			}
			if stat.Kills.Green > 0 {
				greenKD = stat.Kills.Green / stat.Deaths.Green
				currGreenKD = (stat.Kills.Green - startupStats[index].Kills.Green) / (stat.Deaths.Green - startupStats[index].Deaths.Green + 1)
				//fmt.Printf("currk:%6.0f oldk:%6.0f currd:%6.0f oldd:%6.0f\n", stat.Kills.Green, startupStats[index].Kills.Green, stat.Deaths.Green, startupStats[index].Deaths.Green)
			}

			msg += fmt.Sprintf("\nK/D Border %v (%v)\n Blue: %6.1f (%1.1f)\n ", name, stat.Name, dBlue, currBlueKD)
			msg += fmt.Sprintf("Red: %6.1f (%1.1f)\n", redKD, currRedKD)
			msg += fmt.Sprintf("Green: %6.1f (%1.1f)\n", greenKD, currGreenKD)
			//fmt.Println(msg)
		}
		mutex.Lock()
		if len(guilds) > 0 {
			if notteMsg != nil {
				dg.ChannelMessageDelete(guilds[notteTestSrv].Channels[0].ID, notteMsg.ID)
			}
			if sveaUlvarMsg != nil {
				fmt.Println(sveaUlvarMsg.ID)
				dg.ChannelMessageDelete(guilds[sveaUlvarSrv].Channels[0].ID, sveaUlvarMsg.ID)
			}
			notteMsg, _ = dg.ChannelMessageSend(guilds[notteTestSrv].Channels[0].ID, msg)
			sveaUlvarMsg, _ = dg.ChannelMessageSend(guilds[sveaUlvarSrv].Channels[0].ID, msg)

			fmt.Println(msg)
			//fmt.Println(sveaUlvarMsg.ID)
		}
		mutex.Unlock()
		time.Sleep(10 * time.Minute)

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
	botID = u.ID

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)
	dg.AddHandler(guildCreate)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	deleteOldStatsInChannel(notteTestSrv)
	deleteOldStatsInChannel(sveaUlvarSrv)
	//dg.ChannelMessageSend(guilds[1].Channels, "hejsan hoppsan")
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	go runner()
	// Simple way to keep program running until CTRL-C is pressed.
	<-make(chan struct{})

	fmt.Println("Exiting...")
	return
}
