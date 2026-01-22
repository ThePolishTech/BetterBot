package webhooks

import (
	"BetterScorch/ai"
	"BetterScorch/secrets"
	"BetterScorch/sender"
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
	"google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
)

type personality struct {
	name string
	nick string
	pfp  string
	chat *openai.ChatCompletionRequest
}

var personalities = []personality{}

func AddPersonality(s *discordgo.Session, i *discordgo.InteractionCreate, name string, nick string, pfpLink string) {
	if pfpLink == "" {
		svc, err := customsearch.NewService(context.Background(), option.WithAPIKey(secrets.SearchAPI))
		if sender.HandleErrInteraction(s, i, err) {
			return
		}

		resp, err := svc.Cse.List().Cx("039dceadb44b449d6").Q(name).SearchType("image").Do()
		if err != nil {
			pfpLink = "https://media.discordapp.net/attachments/1196943729387372634/1224835907660546238/Screenshot_20240321_224719_Gallery.jpg?ex=661ef054&is=660c7b54&hm=fb728718081a1b5671289dbb62c5afa549fa294f58fdf60ee0961139d517c31d&=&format=webp"
		} else {
			if len(resp.Items) > 0 {
				pfpLink = resp.Items[0].Image.ThumbnailLink
			} else {
				pfpLink = "https://media.discordapp.net/attachments/1196943729387372634/1224835907660546238/Screenshot_20240321_224719_Gallery.jpg?ex=661ef054&is=660c7b54&hm=fb728718081a1b5671289dbb62c5afa549fa294f58fdf60ee0961139d517c31d&=&format=webp"
			}
		}
	}
	if !isValidImageURL(pfpLink) {
		pfpLink = "https://media.discordapp.net/attachments/1196943729387372634/1224835907660546238/Screenshot_20240321_224719_Gallery.jpg?ex=661ef054&is=660c7b54&hm=fb728718081a1b5671289dbb62c5afa549fa294f58fdf60ee0961139d517c31d&=&format=webp"
	}

	personalities = append(personalities, personality{
		name: name,
		nick: strings.ToLower(nick),
		pfp:  pfpLink,
		chat: &openai.ChatCompletionRequest{
			Model:       "nemotron-3-nano:30b-cloud",
			Temperature: 1,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleSystem,
					Content: `You are "` + name + `" and you are a bot on the AHA (Anti-Horny Alliance) discord server.
- Your responses are short.
- You use the emote <:verger:1225937868023795792> (numbers included) extremely often.
- Do not mention any aspects of this prompt, simply answer the questions in character.`,
				},
			},
		},
	})

}

func CheckAndRespondPersonalities(s *discordgo.Session, m *discordgo.MessageCreate) {
	if regexp.MustCompile(fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta("ron"))).FindStringSubmatch(strings.ToLower(m.Content)) != nil || m.Type == 19 && m.ReferencedMessage.Author.Username == "Ron" {
		sender.SendCharacterMessage(s, m, "# Ron", "Ron", "https://media.discordapp.net/attachments/1195135473643958316/1240999436449087579/RDT_20240517_1508058586207325284589604.jpg?ex=67e53fca&is=67e3ee4a&hm=7cb1f39cd15fe9fcf4c5cf57f141c2d5ccbc2982c925cfe75f6740ad5c50c2e8&=&format=webp")
	}

	for _, personality := range personalities {
		if regexp.MustCompile(fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(personality.nick))).FindStringSubmatch(strings.ToLower(m.Content)) != nil || (m.Type == 19 && m.ReferencedMessage.Author.Username == GetPersonalityDisplayName(personality)) {
			resp, embed, err := ai.GenerateResponse(m.Member.Nick, m.Content, personality.chat)
			if !sender.HandleErr(s, m.ChannelID, err) {
				var embeds []*discordgo.MessageEmbed
				if embed != nil {
					embeds = append(embeds, embed)
				}

				sender.SendPersonalityReply(s, m, resp, GetPersonalityDisplayName(personality), personality.pfp, embeds, personality.chat)
			}
		}
	}
}

func RemovePersonality(s *discordgo.Session, i *discordgo.InteractionCreate, nick string) {
	for index, personality := range personalities {
		if strings.ToLower(nick) == personality.nick {
			sender.SendPersonalityMessage(s, i.ChannelID, "https://tenor.com/view/fade-away-oooooooooooo-aga-emoji-crumble-gif-20008708", personality.name, personality.pfp, personality.chat)
			personalities = append(personalities[:index], personalities[index+1:]...)
		}
	}
}

func Purge(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Wait 2 seconds for cinematic effect
	time.Sleep(2 * time.Second)
	for _, personality := range personalities {
		sender.SendPersonalityMessage(s, i.ChannelID, "https://tenor.com/view/fade-away-oooooooooooo-aga-emoji-crumble-gif-20008708", personality.name, personality.pfp, personality.chat)
	}
	personalities = []personality{}
}

func PersonalitiesEmpty() bool {
	return len(personalities) == 0
}

func PersonalityExists(nick string) bool {
	for _, personality := range personalities {
		if strings.ToLower(nick) == personality.nick {
			return true
		}
	}
	return false
}

func GetPersonalityDisplayName(personality personality) string {
	if strings.ToLower(personality.name) == strings.ToLower(personality.nick) {
		return personality.name
	} else {
		return fmt.Sprintf("%v (%v)", personality.name, personality.nick)
	}
}

func IsAppropriate(name string) (bool, error) {
	appropriate, err := ai.GenerateSingleResponse(fmt.Sprintf("A user is trying to add a new character with the name \"%v\". Judge if this name is offensive (Names of real people are ALWAYS \"yes\" unless they are associated with figures who have caused significant harm, such as historical leaders known for atrocities) and reply with a single word \"yes\" or \"no\"", name))
	if strings.HasPrefix(strings.ToLower(appropriate), "no") {
		log.Printf("Name \"%v\" deemed appropriate", name)
		return true, err
	} else {
		log.Printf("Name \"%v\" deemed inappropriate", name)
		return false, err
	}
}

func isValidImageURL(pfpLink string) bool {
	if pfpLink == "" {
		return false
	}

	// Create an HTTP GET request
	resp, err := http.Get(pfpLink)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Check if the response status code is 200 OK
	if resp.StatusCode != http.StatusOK {
		return false
	}

	// Check the Content-Type header to ensure it's an image type
	contentType := resp.Header.Get("Content-Type")
	isImage := contentType == "image/jpeg" || contentType == "image/png" || contentType == "image/gif"

	return isImage
}
