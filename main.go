package main

import (
	"github.com/bwmarrin/discordgo";
	"os";
	"github.com/joho/godotenv";
	"log"
)

func getEnvVariable(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}

var (
	BotToken = getEnvVariable("DISCORD_AUTH_TOKEN")
	ServerIDs = []string{"1234", "5678"}
)

func main() {
	discord,err := discordgo.New("Bot "+ BotToken)
	if err != nil {
		log.Fatalf("error creating Discord session: %v", err)
		return
	}
	discord.AddHandler(messageCreate)
	err = discord.Open()
	if err != nil {
		log.Fatalf("error opening connection: %v", err)
		return
	}
	log.Println("Bot is now running. Press Ctrl+C to exit.")
	select {}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if contains(ServerIDs, m.GuildID) {
		return
	}
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content == "!hello" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Hello, World!")
		if err != nil {
			log.Printf("error sending message : %v", err)
		}
	}
}

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}