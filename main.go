package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	goduckgo "github.com/minoplhy/duckduckgo-images-api"
)

type SearchState struct {
	Results   goduckgo.Gogo
	Index     int
	Timestamp time.Time
	MessageID string
	ChannelID string
}

type HangmanGame struct {
	Mistakes int
	Word string
	MessageID string
	ChannelID string
	CorrectGuesses []string
}

func getEnvVariable(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("error loading .env file")
	}
	return os.Getenv(key)
}

var (
	BotToken     = getEnvVariable("DISCORD_AUTH_TOKEN")
	prefix       = ","
	searchStates = make(map[string]*SearchState)
	statesMutex  = &sync.RWMutex{}
)


// credit to https://gist.github.com/chrishorton/8510732aa9a80a03c829b09f12e20d9c
var HangmanParts = []string{
	`
  +---+
  |   |
      |
      |
      |
      |
=========
	`,
	`
  +---+
  |   |
  O   |
      |
      |
      |
=========
	`,
	`
  +---+
  |   |
  O   |
  |   |
      |
      |
=========
	`,
	`
  +---+
  |   |
  O   |
 /|   |
      |
      |
=========
	`,
	`
  +---+
  |   |
  O   |
 /|\  |
      |
      |
=========
	`,
	`
  +---+
  |   |
  O   |
 /|\  |
 /    |
      |
=========
	`,
	`
  +---+
  |   |
  O   |
 /|\  |
 / \  |
      |
=========
	`,
}



func main() {
	discord, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("error creating Discord session: %v", err)
		return
	}
	discord.AddHandler(messageCreate)
	discord.AddHandler(interactionsCreate)

	err = discord.Open()
	if err != nil {
		log.Fatalf("error opening connection: %v", err)
		return
	}
	log.Println("Bot is now running. Press Ctrl+C to exit.")

	go cleanupRoutine()

	select {}
}

func cleanupRoutine() {
	for {
		time.Sleep(10 * time.Minute)

		now := time.Now()

		statesMutex.Lock()
		for messageID, state := range searchStates {
			if now.Sub(state.Timestamp) > 15*time.Minute {
				delete(searchStates, messageID)
			}
		}
		statesMutex.Unlock()
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	user, err := s.User(m.Author.ID)
	if err != nil {
		log.Printf("error retrieving user: %v", err)
	}

	if m.Content == prefix+"hello" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Hello, <@"+user.ID+">!")
		if err != nil {
			log.Printf("error sending message : %v", err)
		}
	}

	if m.Content == prefix+"userinfo" {
		embed := &discordgo.MessageEmbed{
			Title: "User Info",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Username:",
					Value:  user.Username,
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

	if strings.HasPrefix(m.Content, prefix+"s") {
		splitSearch := strings.SplitN(m.Content, " ", 2)
		if len(splitSearch) == 1 {
			s.ChannelMessageSend(m.ChannelID, "Please provide a search query.")
			log.Printf("error, no search query: %v", m.Content)
			return
		}

		searchQuery := splitSearch[1]

		imageSearch := goduckgo.Search((goduckgo.Query{Keyword: searchQuery}))

		if len(imageSearch.Results) == 0 {
			s.ChannelMessageSend(m.ChannelID, "No results found.")
			return
		}

		embed := &discordgo.MessageEmbed{
			Color: 0xFFFFFF,
			Author: &discordgo.MessageEmbedAuthor{
				Name:    imageSearch.Results[0].Title,
				URL:     imageSearch.Results[0].URL,
				IconURL: imageSearch.Results[0].Thumbnail,
			},
			Image: &discordgo.MessageEmbedImage{
				URL: imageSearch.Results[0].Image,
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "(🤖) DuckDuckGo Search | Page 1/" + strconv.Itoa(len(imageSearch.Results)),
			},
		}
		buttons := []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.Button{
						Style:    discordgo.SecondaryButton,
						Label:    "⬅️",
						CustomID: "img_backward",
					},
					&discordgo.Button{
						Style:    discordgo.SecondaryButton,
						Label:    "➡️",
						CustomID: "img_forward",
					},
				},
			},
		}
		msg, err := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Embed:     embed,
			Components: buttons,
		})
		if err != nil {
			log.Printf("error sending search results: %v", err)
			return
		}
		statesMutex.Lock()
		searchStates[msg.ID] = &SearchState{
			Results:   imageSearch,
			Index:     0,
			Timestamp: time.Now(),
			MessageID: msg.ID,
			ChannelID: m.ChannelID,
		}
		statesMutex.Unlock()
		go func(messageID string) {
			time.Sleep(15 * time.Minute)
			statesMutex.Lock()
			delete(searchStates, messageID)
			statesMutex.Unlock()
		}(msg.ID)
	}
	if strings.HasPrefix(m.Content, prefix+"hangman") {

	}

}

func interactionsCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}
	msgID := i.Message.ID

	statesMutex.RLock()
	state, exists := searchStates[msgID]
	statesMutex.RUnlock()

	if !exists {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "This session has expired. Please initiate a new search.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Println("error responding to interaction,", err)
		}
		return
	}

	
	switch i.MessageComponentData().CustomID {
	case "img_forward":
		state.Index = (state.Index + 1) % len(state.Results.Results)
	case "img_backward":
		state.Index = (state.Index - 1 + len(state.Results.Results)) % len(state.Results.Results)
	}

	result := state.Results.Results[state.Index]
	embed := &discordgo.MessageEmbed{
		Color: 0xFFFFFF,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    result.Title,
			URL:     result.URL,
			IconURL: result.Thumbnail,
		},
		Image: &discordgo.MessageEmbedImage{
			URL: result.Image,
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "(🤖) DuckDuckGo Search | Page " + strconv.Itoa(state.Index+1) + "/" + strconv.Itoa(len(state.Results.Results)),
		},
	}

	
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Println("error responding to interaction,", err)
		return
	}

	
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					Style:    discordgo.SecondaryButton,
					Label:    "⬅️",
					CustomID: "img_backward",
				},
				&discordgo.Button{
					Style:    discordgo.SecondaryButton,
					Label:    "➡️",
					CustomID: "img_forward",
				},
			},
		},
	}

	
	edit := &discordgo.MessageEdit{
		ID:         state.MessageID,
		Channel:    state.ChannelID,
		Embed:      embed,
		Components: &components, 
	}
	_, err = s.ChannelMessageEditComplex(edit)
	if err != nil {
		log.Println("Error editing message:", err)
	}
}



