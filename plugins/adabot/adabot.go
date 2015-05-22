package adabot

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/kyokomi/slackbot/plugins"
	"github.com/kyokomi/slackbot/slackctx"
	"golang.org/x/net/context"
)

type pluginKey string

func init() {
	module := AdaMessage{}
	plugins.AddPlugin(pluginKey("adaMessage"), module)

	fmt.Println("Got to init")
	db, err := bolt.Open("ada.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	err2 := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("definitions"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	if err2 != nil {
		log.Fatal(err2)
	}

	defer db.Close()

}

var rd = rand.New(rand.NewSource(time.Now().UnixNano()))

// Message is Slack Receive Message.
type Message struct {
	userID, userName, channelID, channelName, text string
}

func NewMessage(userID, channelID, text string) Message {
	var m Message
	m.userID = userID
	m.channelID = channelID
	m.text = text
	return m
}

type AdaMessage struct {
}

func (r AdaMessage) CheckMessage(ctx context.Context, message string) (bool, string) {

	db, dberr := bolt.Open("ada.db", 0600, nil)
	if dberr != nil {
		log.Fatal(dberr)
	}
	defer db.Close()

	api := slackctx.FromSlackClient(ctx)
	botUser := api.GetInfo().User
	botName := api.Name

	if strings.Index(message, botUser.Id) != -1 {
		// If the message has @ada in it somewhere
		message = message[strings.Index(message, ":")+len(":"):]
	} else {
		return false, ""
	}

	re, _ := regexp.Compile("define (.*) as (.*)")
	if re.MatchString(message) {
		submatches := re.FindAllStringSubmatch(message, 2)
		fmt.Println(submatches[0][1] + "->" + submatches[0][2])
		term := submatches[0][1]
		definition := submatches[0][2]

		err := db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("definitions"))
			err := b.Put([]byte(term), []byte(definition))
			return err
		})

		if err != nil {
			return true, "Something freaky just happened. I can't define anything right now. Call @thorinside and tell him " + err.Error() + "!!!"
		}

		return true, "Alright, '" + term + "' is " + definition
	}

	re2, _ := regexp.Compile("what is (?:a )?(.*)\\?")
	re3, _ := regexp.Compile("what does (.*) mean\\?")
	var submatches [][]string
	var found bool
	if re2.MatchString(message) {
		submatches = re2.FindAllStringSubmatch(message, 1)
		found = true
	} else if re3.MatchString(message) {
		submatches = re3.FindAllStringSubmatch(message, 1)
		found = true
	}

	if found {
		term := submatches[0][1]

		var definition string

		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("definitions"))
			value := b.Get([]byte(term))
			if value != nil {
				definition = string(value[:])
			}
			return nil
		})

		if err != nil {
			return true, "Something freaky just happened. I can't forget anything right now. Call @thorinside and tell him " + err.Error() + "!!!"
		}

		if definition != "" {
			return true, "According to my records, '" + term + "' is defined as " + definition
		} else {
			return true, "Sorry, '" + term + "' is undefined."
		}
	}

	re4, _ := regexp.Compile("forget (.*)")
	if re4.MatchString(message) {
		submatches := re4.FindAllStringSubmatch(message, 1)
		term := submatches[0][1]

		err := db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("definitions"))
			err := b.Delete([]byte(term))
			return err
		})

		if err != nil {
			return true, "Something freaky just happened. I can't forget anything right now. Call @thorinside and tell him " + err.Error() + "!!!"
		}
		return true, "Okay, forgetting '" + term + "'."
	}

	return true, message
}

func (r AdaMessage) DoAction(ctx context.Context, message string) bool {
	// msEvent := slackctx.FromMessageEvent(ctx)

	//m := NewMessage(msEvent.UserId, msEvent.ChannelId, message)
	plugins.SendMessage(ctx, message)
	return true
}

var _ plugins.BotMessagePlugin = (*AdaMessage)(nil)
