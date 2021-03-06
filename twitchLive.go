package main

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/go-ini/ini"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
	"strings"
)

type Stream struct {
	StreamData       []Data     `json:"data"`
	StreamPagination Pagination `json:"pagination"`
}

type Data struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	GameID       string    `json:"game_id"`
	CommunityIds []string  `json:"community_ids"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	ViewerCount  int       `json:"viewer_count"`
	StartedAt    time.Time `json:"started_at"`
	Language     string    `json:"language"`
	ThumbnailURL string    `json:"thumbnail_url"`
}

type Pagination struct {
	Cursor string `json:"cursor"`
}

type User struct {
	UserData []UsersData `json:"data"`
}

type UsersData struct {
	BroadcasterType string `json:"broadcaster_type"`
	Description     string `json:"description"`
	DisplayName     string `json:"display_name"`
	Email           string `json:"email"`
	ID              string `json:"id"`
	Login           string `json:"login"`
	OfflineImageURL string `json:"offline_image_url"`
	ProfileImageURL string `json:"profile_image_url"`
	Type            string `json:"type"`
	ViewCount       int    `json:"view_count"`
}

type Game struct {
	Data []GameData `json:"data"`
}

type GameData struct {
	ID        string `json:"id"`
	BoxArtURL string `json:"box_art_url"`
	Name      string `json:"name"`
}

var twitchKey string

// Generated by curl-to-Go: https://mholt.github.io/curl-to-go
//func twitchLive(s *discordgo.Session, msg *discordgo.MessageCreate, arg string){
func twitchLive(arg string) {

	req, err := http.NewRequest("GET", "https://api.twitch.tv/helix/streams?user_login="+arg, nil)
	if err != nil {
		fmt.Println("SHIT ", err)
	}
	req.Header.Set("Client-Id", twitchKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("SHIT2", err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close()

	if err != nil {
		log.Fatal("TwitchAPI error:", err)
	}

	stream, err := getStreams([]byte(body))

	if len(stream.StreamData) > 0 {
		req2, err := http.NewRequest("GET", "https://api.twitch.tv/helix/users?login="+arg, nil)
		if err != nil {
			fmt.Println("SHIT ", err)
		}
		req2.Header.Set("Client-Id", twitchKey)

		resp2, err := http.DefaultClient.Do(req2)
		if err != nil {
			fmt.Println("SHIT2", err)
		}

		body2, err := ioutil.ReadAll(resp2.Body)

		defer resp2.Body.Close()

		if err != nil {
			log.Fatal("TwitchAPI error:", err)
		}

		user, err := getUsers([]byte(body2))

		req3, err := http.NewRequest("GET", "https://api.twitch.tv/helix/games?id="+stream.StreamData[0].GameID, nil)
		if err != nil {
			fmt.Println("SHIT ", err)
		}
		req3.Header.Set("Client-Id", twitchKey)

		resp3, err := http.DefaultClient.Do(req3)
		if err != nil {
			fmt.Println("SHIT2", err)
		}

		body3, err := ioutil.ReadAll(resp3.Body)

		defer resp3.Body.Close()

		if err != nil {
			log.Fatal("TwitchAPI error:", err)
		}

		game, err := getGames([]byte(body3))

		current := time.Now().Unix()
		start := stream.StreamData[0].StartedAt.Unix()
		if stream.StreamData[0].Type == "live" && current-start >= 300 && current-start < 600{
			//fmt.Println(stream.StreamData[0].Type)

			embed := &discordgo.MessageEmbed{
				URL: "https://twitch.tv/" + arg,
				Author: &discordgo.MessageEmbedAuthor{
					URL:  "https://twitch.tv/" + arg,
					Name: user.UserData[0].DisplayName,
				},
				Color: 0x00ff00, // Green
				Fields: []*discordgo.MessageEmbedField{
					&discordgo.MessageEmbedField{
						Name:   "Playing",
						Value:  game.Data[0].Name,
						Inline: true,
					},
				},
				Image: &discordgo.MessageEmbedImage{
					URL: "https://static-cdn.jtvnw.net/previews-ttv/live_user_" + strings.ToLower(arg) + "-320x180.jpg&time=" + strconv.FormatInt(current, 10),
				},
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: user.UserData[0].ProfileImageURL,
				},
				//Timestamp: time.Now().String(),
				Title: stream.StreamData[0].Title,
			}
			//fmt.Println(stream.StreamData[0].Type)
			_, err := dgSession.ChannelMessageSendEmbed("362408790051651597", embed)
			if err != nil {
				fmt.Println(err)
			}

		}
	}
	//fmt.Println(stream.Data)
}

func getStreams(body []byte) (*Stream, error) {
	var s = new(Stream)
	err := json.Unmarshal(body, &s)
	if err != nil {
		fmt.Println("whoops:", err)
	}
	return s, err
}

func getUsers(body []byte) (*User, error) {
	var s = new(User)
	err := json.Unmarshal(body, &s)
	if err != nil {
		fmt.Println("whoops:", err)
	}
	return s, err
}

func getGames(body []byte) (*Game, error) {
	var s = new(Game)
	err := json.Unmarshal(body, &s)
	if err != nil {
		fmt.Println("whoops:", err)
	}
	return s, err
}

func addStream(s *discordgo.Session, msg *discordgo.MessageCreate, stream string) {
	newItem := "INSERT INTO streams (TwitchPage,DiscorduID) values (?,?)"
	stmt, err := db.Prepare(newItem)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(stream, msg.Author.ID)
	if err2 != nil {
		panic(err2)
	}
	s.ChannelMessageSend(msg.ChannelID, "Added your stream to the database:`https://www.twitch.tv/"+stream+"`")

}

func checkDB() {
	for true {
		if hasSession {
			qte, err := db.Query("SELECT TwitchPage FROM streams")
			if err != nil {
				log.Fatal("Query error:", err)
			}
			defer qte.Close()

			var stream string
			for qte.Next() {
				err = qte.Scan(&stream)
				if err != nil {
					log.Fatal("Parse error:", err)
				}
				twitchLive(stream)
			}

			time.Sleep(300000 * time.Millisecond)
		}
	}

}

func twitchLiveTester(s *discordgo.Session, msg *discordgo.MessageCreate, arg string) {
	req, err := http.NewRequest("GET", "https://api.twitch.tv/helix/streams?user_login="+arg, nil)
	if err != nil {
		fmt.Println("Error creating twitch request:", err)
	}
	req.Header.Set("Client-Id", twitchKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending twitch request:", err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close()

	if err != nil {
		log.Fatal("TwitchAPI error:", err)
	}

	stream, err := getStreams([]byte(body))

	if len(stream.StreamData) > 0 {
		req2, err := http.NewRequest("GET", "https://api.twitch.tv/helix/users?login="+arg, nil)
		if err != nil {
			fmt.Println("SHIT ", err)
		}
		req2.Header.Set("Client-Id", twitchKey)

		resp2, err := http.DefaultClient.Do(req2)
		if err != nil {
			fmt.Println("SHIT2", err)
		}

		body2, err := ioutil.ReadAll(resp2.Body)

		defer resp2.Body.Close()

		if err != nil {
			log.Fatal("TwitchAPI error:", err)
		}

		user, err := getUsers([]byte(body2))

		req3, err := http.NewRequest("GET", "https://api.twitch.tv/helix/games?id="+stream.StreamData[0].GameID, nil)
		if err != nil {
			fmt.Println("SHIT ", err)
		}
		req3.Header.Set("Client-Id", twitchKey)

		resp3, err := http.DefaultClient.Do(req3)
		if err != nil {
			fmt.Println("SHIT2", err)
		}

		body3, err := ioutil.ReadAll(resp3.Body)

		defer resp3.Body.Close()

		if err != nil {
			log.Fatal("TwitchAPI error:", err)
		}

		game, err := getGames([]byte(body3))

		current := time.Now().Unix()
		//start:= stream.StreamData[0].StartedAt.Unix()
		if stream.StreamData[0].Type == "live" { //&& current-start < 300{
			//fmt.Println(stream.StreamData[0].Type)

			embed := &discordgo.MessageEmbed{
				URL: "https://twitch.tv/" + arg,
				Author: &discordgo.MessageEmbedAuthor{
					URL:  "https://twitch.tv/" + arg,
					Name: user.UserData[0].DisplayName,
				},
				Color: 0x00ff00, // Green
				Fields: []*discordgo.MessageEmbedField{
					&discordgo.MessageEmbedField{
						Name:   "Playing",
						Value:  game.Data[0].Name,
						Inline: true,
					},
				},
				Image: &discordgo.MessageEmbedImage{
					URL: "https://static-cdn.jtvnw.net/previews-ttv/live_user_" + strings.ToLower(arg) + "-320x180.jpg&time=" + strconv.FormatInt(current, 10),
				},
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: user.UserData[0].ProfileImageURL,
				},
				//Timestamp: time.Now().String(),
				Title: stream.StreamData[0].Title,
			}
			//fmt.Println(stream.StreamData[0].Type)
			_, err := dgSession.ChannelMessageSendEmbed("379073357401948162", embed)
			if err != nil {
				fmt.Println(err)
			}

		}
		//fmt.Println(stream.StreamData[0].Type)
	}
	//fmt.Println(stream.Data)
}

func init() {
	CmdList["twitchtest"] = twitchLiveTester
	CmdList["addstream"] = addStream
	cfg, err := ini.Load("./config/truebot.ini")
	if err != nil {
		fmt.Println("Was not able to load Twitch API Key - ", err)
	}
	twitchKey = cfg.Section("api-keys").Key("twitch").String()
	go checkDB()
}

//TODO
/*

 */
