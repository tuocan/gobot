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
	prefix = "!"
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
	user, err := s.User(m.Author.ID)
	if err != nil {
		log.Printf("error retrieving user: %v", err)
	}
	if m.Content == prefix+"hello" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Hello, World!")
		if err != nil {
			log.Printf("error sending message : %v", err)
		}
	}
	if m.Content == prefix+"userinfo" {
		embed := &discordgo.MessageEmbed{
			Title: "User Info",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name: "Username:",
					Value: user.Username,
					Inline: false,
				},
				{
					Name:   "User ID:",
					Value:  user.ID,
					Inline: false,
				},
				{
					Name:   "Global Name:",
					Value:  user.GlobalName,
					Inline: false,
				},
			},
            Thumbnail: &discordgo.MessageEmbedThumbnail{
                URL: user.AvatarURL("2048"),
            },
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)
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