package main

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/garyburd/redigo/redis"
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
	c, err := redis.DialURL(os.Getenv("REDISTOGO_URL"))
	if err != nil {
		return "", err
	}
	defer c.Close()

	words := strings.Split(msg.Text, " ")[1:]

	if len(words) == 0 {
		return "", errors.New("invalid usage")
	}

	switch words[0] {
	case "get":
		url, err := redis.String(c.Do("GET", msg.User))
		if err == redis.ErrNil {
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
		_, err := c.Do("SET", msg.User, url)
		if err != nil {
			return "", err
		}
		return ":dizzy: URL - " + url, nil
	case "summary":
		url, err := redis.String(c.Do("GET", msg.User))
		if err == redis.ErrNil {
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
