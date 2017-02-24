package main

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"gopkg.in/redis.v5"

	"github.com/nlopes/slack"
)

const (
	list = "list_"
)

func run(api *slack.Client) int {
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				if strings.Split(ev.Msg.Text, " ")[0] == "linka" {
					log.Println(ev.Msg.Text)
					msg, err := response(ev.Msg)
					if err != nil {
						log.Println(err)
					}
					rtm.SendMessage(rtm.NewOutgoingMessage(msg, ev.Channel))
				}

			case *slack.InvalidAuthEvent:
				log.Println("invalid cred")
				return 1
			}
		}
	}
}

func response(msg slack.Msg) (string, error) {
	client := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDISTOGO_URL"),
	})
	defer client.Close()

	words := strings.Split(msg.Text, " ")[1:]

	if len(words) == 0 {
		return "", errors.New("invalid usage")
	}

	switch words[0] {
	case "get":
		url, err := client.Get(msg.User).Result()
		if err == redis.Nil {
			return "", errors.New("registered user's url not found")
		} else if err != nil {
			return "", err
		}
		return ":dizzy: URL - " + url, nil
	case "set":
		if len(words) == 1 {
			return "", errors.New("invalid usage")
		}
		url := words[1]
		url = url[1 : len(url)-1]
		err := client.Set(msg.User, url, 0).Err()
		if err != nil {
			return "", err
		}
		return ":dizzy: URL - " + url, nil
	case "summary":
		url, err := client.Get(msg.User).Result()
		if err == redis.Nil {
			return "", errors.New("registered user's url not found")
		} else if err != nil {
			return "", err
		}
		return summary(url), nil
	default:
		return "", errors.New("invalid usage")
	}
}

func summary(url string) string {
	res, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer res.Body.Close()

	text, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return ""
	}

	return string(text)
}

func main() {
	api := slack.New(os.Getenv("SLACK_TOKEN"))
	os.Exit(run(api))
}
