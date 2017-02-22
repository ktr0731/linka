package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/nlopes/slack"
)

func run(api *slack.Client) int {
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				if strings.HasPrefix(ev.Msg.Text, "linka ") || ev.Msg.Text == "linka" {
					log.Println(ev.Msg.Text)
					rtm.SendMessage(rtm.NewOutgoingMessage(message(), ev.Channel))
				}

			case *slack.InvalidAuthEvent:
				log.Println("invalid cred")
				return 1
			}
		}
	}
}

func message() string {
	res, err := http.Get(os.Getenv("URL"))
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
