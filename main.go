package main

import (
	"log"
	"os"
	"strings"
	"time"
	// "strconv"
	"github.com/pborman/uuid"
	tb "gopkg.in/tucnak/telebot.v2"
)

type viewer struct {
	Name     string
	Username string
	Ready    bool
	Paused   bool
}

type watchParty struct {
	ID          string
	Name        string
	Viewers     []viewer
	Step        string
	ChatID      int64
	OwnerID     int
	ReadyMsgID  int
	PausedMsgID int
}

type replyID struct {
	ChatID int64
	MsgID int
}

var data []*watchParty
var replyIDs []*replyID

func main() {

	var (
		port      = os.Getenv("PORT")
		publicURL = os.Getenv("PUBLIC_URL") // you must add it to your config vars
		token     = os.Getenv("TOKEN")      // you must add it to your config vars
	)

	webhook := &tb.Webhook{
		Listen:   ":" + port,
		Endpoint: &tb.WebhookEndpoint{PublicURL: publicURL},
	}

	pref := tb.Settings{
		Token:  token,
		Poller: webhook,
	}

	b, err := tb.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	var (
		// Universal markup builders.
		menu     = &tb.ReplyMarkup{ResizeReplyKeyboard: true}
		selector = &tb.ReplyMarkup{}

		// Reply buttons.
		btnHelp     = menu.Text("ℹ Help")
		btnSettings = menu.Text("⚙ Settings")

		// Inline buttons.
		//
		// Pressing it will cause the client to
		// send the bot a callback.
		//
		// Make sure Unique stays unique as per button kind,
		// as it has to be for callback routing to work.
		//
		btnPrev = selector.Data("⬅", "prev", "1")
		btnNext = selector.Data("➡", "next", "1")
	)


	menu.Reply(
		menu.Row(btnHelp),
		menu.Row(btnSettings),
	)
	selector.Inline(
		selector.Row(btnPrev, btnNext),
	)

	replyquery := &tb.ReplyMarkup{ForceReply: true}

	b.Handle("/start", func(m *tb.Message) {
		filmName := m.Text[6:]
		filmName = strings.TrimSpace(filmName)
		if len(filmName) == 0 {
			rep, _ := b.Send(m.Chat, "Please specify film or show name:", replyquery)
			addNewReplyId(m.Chat.ID, rep.ID)
			return
		}
		b.Send(m.Chat, "OK! Setting us up to watch "+filmName)
		handleNewWatchParty(b, filmName, m.Chat.ID, m.Sender.ID, m.Chat)
	})

	b.Handle(tb.OnText, func(m *tb.Message) {
		if m.ReplyTo != nil && checkReplyIDExists(m.Chat.ID, m.ReplyTo.ID) {
			deleteReplyID(m.Chat.ID, m.ReplyTo.ID)
			filmName := strings.TrimSpace(m.Text)
			if (len(filmName) == 0) {
				return
			}
			b.Send(m.Chat, "OK! Setting us up to watch " + filmName)
			handleNewWatchParty(b, filmName, m.Chat.ID, m.Sender.ID, m.Chat)
		}
	})

	// On reply button pressed (message)
	b.Handle(&btnHelp, func(m *tb.Message) {})

	// On inline button pressed (callback)
	b.Handle(&btnPrev, func(c *tb.Callback) {
		// ...
		// Always respond!
		b.Respond(c, &tb.CallbackResponse{Text: "Previous"})
	})

	b.Handle(&btnNext, func(c *tb.Callback) {
		// ...
		// Always respond!
		b.Respond(c, &tb.CallbackResponse{Text: "Next"})
	})

	b.Handle("/count", func(m *tb.Message) {
		b.Send(m.Chat, "3")
		time.Sleep(1 * time.Second)
		b.Send(m.Chat, "2")
		time.Sleep(1 * time.Second)
		b.Send(m.Chat, "1")
		time.Sleep(1 * time.Second)
		b.Send(m.Chat, "Go!")
	})

	b.Handle("/playstation", func(m *tb.Message) {
		b.Send(m.Chat, "P")
		time.Sleep(150 * time.Millisecond)
		b.Send(m.Chat, "L")
		time.Sleep(150 * time.Millisecond)
		b.Send(m.Chat, "A")
		time.Sleep(150 * time.Millisecond)
		b.Send(m.Chat, "Y")
		time.Sleep(150 * time.Millisecond)
		b.Send(m.Chat, "S")
		time.Sleep(150 * time.Millisecond)
		b.Send(m.Chat, "T")
		time.Sleep(150 * time.Millisecond)
		b.Send(m.Chat, "A")
		time.Sleep(150 * time.Millisecond)
		b.Send(m.Chat, "T")
		time.Sleep(150 * time.Millisecond)
		b.Send(m.Chat, "I")
		time.Sleep(150 * time.Millisecond)
		b.Send(m.Chat, "O")
		time.Sleep(150 * time.Millisecond)
		b.Send(m.Chat, "N")
		time.Sleep(150 * time.Millisecond)
	})

	b.Handle("/llama", func(m *tb.Message) {
		a := &tb.Photo{File: tb.FromURL("https://pbs.twimg.com/profile_images/378800000802823295/fa4f4104d718899ea49f3a507c7f6034_400x400.jpeg")}
		if err != nil {
			return
		}
		b.Send(m.Chat, a)
	})

	b.Handle("/randomllama", func(m *tb.Message) {
		a := &tb.Photo{File: tb.FromURL("https://source.unsplash.com/800x600?llama")}
		if err != nil {
			return
		}
		b.Send(m.Chat, a)
	})

	b.Start()
}

func addNewReplyId(chatID int64, msgID int) {
	replyIDs = append(replyIDs, &replyID{ChatID: chatID, MsgID: msgID})
}

func checkReplyIDExists(chatID int64, msgID int) bool {
	for _, r := range replyIDs {
		if chatID == r.ChatID && msgID == r.MsgID {
			return true
		}
	}
	return false
}

func deleteReplyID(chatID int64, msgID int) bool {
	for i, r := range replyIDs {
		if chatID == r.ChatID && msgID == r.MsgID {
			replyIDs = append(replyIDs[:i], replyIDs[i+1:]...)
			return true
		}
	}
	return false
}

func getWatchPartyByID(ID string) *watchParty {
	for _, wp := range data {
		if ID == wp.ID {
			return wp
		}
	}
	return nil
}

func createNewWatchParty(name string, chatID int64, ownerID int) string {
	id := uuid.New()
	wp := &watchParty{ID: id, Name: name, ChatID: chatID, OwnerID: ownerID}
	data = append(data, wp)
	return id
}

func handleNewWatchParty(b *tb.Bot, filmName string, chatId int64, senderID int, chat *tb.Chat) {
	wpID := createNewWatchParty(filmName, chatId, senderID)

	InOrOut := &tb.ReplyMarkup{}
	btnIn := InOrOut.Data("I'm in!", "in", wpID)
	btnOut := InOrOut.Data("I'm not in", "out", wpID)
	InOrOut.Inline(InOrOut.Row(btnIn, btnOut))

	b.Handle(&btnIn, func(c *tb.Callback) {
		b.Respond(c, &tb.CallbackResponse{Text: "Noted that you are in!"})
	})
	b.Handle(&btnOut, func(c *tb.Callback) {
		b.Respond(c, &tb.CallbackResponse{Text: c.Data})
	})
	b.Send(chat, "Who's in?", InOrOut)

	readyNotReady := &tb.ReplyMarkup{}
	btnReady := InOrOut.Data("Paused and Ready!", "ready", wpID)
	btnNotReady := InOrOut.Data("Not ready!", "notready", wpID)
	readyNotReady.Inline(InOrOut.Row(btnReady, btnNotReady))

	b.Handle(&btnReady, func(c *tb.Callback) {
		b.Respond(c, &tb.CallbackResponse{Text: "Noted that you are ready!"})
	})
	b.Handle(&btnNotReady, func(c *tb.Callback) {
		b.Respond(c, &tb.CallbackResponse{Text: "Noted that you are not ready"})
	})

	b.Send(chat, "Please pause at 3 seconds, when you are ready hit ready. Ready status will last for 30 seconds. \n\n Not Ready: \n\n Ready: \n Max", btnNotReady)
}
