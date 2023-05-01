package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

const prefix string = "!gobot"

type Answers struct {
	OriginChannelId string
	FavCharacter    string
	FavGame         string
	RecordId        int64
}

func (a *Answers) ToMessageEmbed() *discordgo.MessageEmbed {
	fields := []*discordgo.MessageEmbedField{
		{
			Name:  "Favorite Character",
			Value: a.FavCharacter,
		},
		{
			Name:  "Favorite Game",
			Value: a.FavGame,
		},
		{
			Name:  "Record ID",
			Value: strconv.FormatInt(a.RecordId, 10),
		},
	}
	return &discordgo.MessageEmbed{
		Title:  "New Responses",
		Fields: fields,
	}
}

var responses map[string]Answers = map[string]Answers{}

func main() {
	godotenv.Load()
	fmt.Println("Starting Discord Bot")
	token := os.Getenv("BOT_TOKEN")

	log.Println("Token: ", token)
	sess, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sess.AddHandler(func(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
		if r.Emoji.Name == "🔥" {
			s.GuildMemberRoleAdd(r.GuildID, r.UserID, "852000000000000000")
			s.ChannelMessageSend(r.ChannelID, fmt.Sprintf("%v has been added to %v", r.UserID, r.Emoji.Name))
		}
	})

	sess.AddHandler(func(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
		if r.Emoji.Name == "🔥" {
			s.GuildMemberRoleAdd(r.GuildID, r.UserID, "852000000000000000")
			s.ChannelMessageSend(r.ChannelID, fmt.Sprintf("%v has been removed from %v", r.UserID, r.Emoji.Name))
		}
	})

	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}
		if m.GuildID == "" {
			UserPromptResponseHandler(db, s, m)
		}
		args := strings.Split(m.Content, " ")
		if args[0] != prefix {
			return
		}

		if args[1] == "Hello" {
			s.ChannelMessageSend(m.ChannelID, "Fuck you and welcome to Omar's bot")
		}

		if args[1] == "batman" {
			proverbs := []string{
				"Oh, you think darkness is your ally. But you merely adopted the dark; I was born in it, moulded by it.",
				"I will show you where I have made my home while preparing to bring justice. Then I will break you.",
				"Peace has cost you your strength. Victory has defeated you.",
				"The shadows betray you, because they belong to me!",
				"Introduce a little anarchy. Upset the established order and everything becomes chaos.",
				"Do I really look like a guy with a plan?",
				"Some men aren't looking for anything logical, like money. They can't be bought, bullied, reasoned, or negotiated with. Some men just want to watch the world burn.",
				"He's a madman, but he's also a genius. He's a man with no rules and no boundaries, and that makes him dangerous.",
				"The Joker is a master of chaos and deception. He'll do whatever it takes to get what he wants, and he'll never stop until he's destroyed everything in his path.",
			}

			selection := rand.Intn(len(proverbs))
			author := discordgo.MessageEmbedAuthor{
				Name: "The Dark Knight",
				URL:  "https://www.youtube.com/watch?v=EXeTwQWrcwY",
			}
			embed := discordgo.MessageEmbed{
				Title:  proverbs[selection],
				Author: &author,
			}
			s.ChannelMessageSendEmbed(m.ChannelID, &embed)
		}

		if args[1] == "motivate" {
			file, err := os.Open("Quotes.txt")
			if err != nil {
				fmt.Println(err)
				return
			}
			defer file.Close()

			// Read the file line by line and add each line to the array
			proverbs := []string{}
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				proverbs = append(proverbs, scanner.Text())
			}
			selection3 := rand.Intn(len(proverbs))
			author := discordgo.MessageEmbedAuthor{
				Name: "Man In The Mirror",
				URL:  "https://www.youtube.com/@Caleb_Duplain",
			}
			embed3 := discordgo.MessageEmbed{
				Title:  proverbs[selection3],
				Author: &author,
			}
			s.ChannelMessageSendEmbed(m.ChannelID, &embed3)
		}

		if args[1] == "podcast" {
			file, err := os.Open("Podcasts.txt")
			if err != nil {
				fmt.Println(err)
				return
			}
			defer file.Close()

			// Read the file line by line and add each line to the array
			proverbs := []string{}
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				proverbs = append(proverbs, scanner.Text())
			}
			selection3 := rand.Intn(len(proverbs))
			s.ChannelMessageSend(m.ChannelID, "/play "+proverbs[selection3])
		}

		if args[1] == "prompt" {
			UserPromptHandler(s, m)
		}

		if args[1] == "answer" {
			AnswerHandler(db, s, m)
		}

		if args[1] == "help" {
			embed := discordgo.MessageEmbed{
				Title: "Help",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Commands",
						Value:  "insult, compliment, proverbs, prompt",
						Inline: false,
					},
					{
						Name:   "Usage",
						Value:  "insult, compliment, proverbs, prompt",
						Inline: false,
					},
				},
			}
			s.ChannelMessageSendEmbed(m.ChannelID, &embed)
		}
	})

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = sess.Open()

	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()
	fmt.Println("Bot is working my G")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

}

func AnswerHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {
	spl := strings.Split(m.Content, " ")
	if len(spl) < 3 {
		s.ChannelMessageSend(m.ChannelID, "An id must be provided, Ex :'!gobot answer 1' ")
		return
	}
	id, err := strconv.Atoi(spl[2])
	if err != nil {
		log.Panic(err)
	}
	var recordId int64
	var answerStr string
	var userId int64
	query := "select * from discord_messages where id = ?"
	row := db.QueryRow(query, id)
	err = row.Scan(&recordId, &answerStr, &userId)
	if err != nil {
		log.Panic(err)
	}

	var answers Answers
	err = json.Unmarshal([]byte(answerStr), &answers)
	if err != nil {
		log.Panic(err)
	}
	answers.RecordId = recordId
	embed := answers.ToMessageEmbed()
	s.ChannelMessageSendEmbed(m.ChannelID, embed)

}

func UserPromptHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	channel, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		log.Panic(err)
	}
	if _, ok := responses[channel.ID]; !ok {
		responses[channel.ID] = Answers{
			OriginChannelId: m.ChannelID,
			FavCharacter:    "",
			FavGame:         "",
		}
		s.ChannelMessageSend(channel.ID, "Here are some questions homeboy : ")
		s.ChannelMessageSend(channel.ID, "What is your favorite character from the Dark Knight Trilogy?")
	} else {
		s.ChannelMessageSend(m.ChannelID, "You retarded or sum bruv... we still waiting (ENGLISH MOTHERFUCKER )")
	}
}

func UserPromptResponseHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {
	answer, ok := responses[m.ChannelID]
	if !ok {
		return
	}
	if answer.FavCharacter == "" {
		answer.FavCharacter = m.Content
		s.ChannelMessageSend(m.ChannelID, "What is your favorite game then boiii ?")
		responses[m.ChannelID] = answer
		return
	} else if answer.FavGame == "" {
		answer.FavGame = m.Content

		query := "insert into discord_messages (payload , user_id) values(?,?)"
		jbytes, err := json.Marshal(answer)
		if err != nil {
			log.Fatal(err)
		}

		res, err := db.Exec(query, string(jbytes), m.ChannelID)
		lastInserted, err := res.LastInsertId()
		if err != nil {
			log.Fatal(err)
		}
		answer.RecordId = lastInserted

		embed := answer.ToMessageEmbed()
		s.ChannelMessageSendEmbed(answer.OriginChannelId, embed)
		delete(responses, m.ChannelID)
	}
}
