package bot

import (
	"bargo/constants"
	"bargo/storage"
	"time"
	"log"
	"bargo/logic"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	tb "github.com/AntoniBertel/telebot"
)

var bot *tb.Bot

func Start() {
	b, err := tb.NewBot(tb.Settings{
		Token:  constants.OutsideConstants.TelegramBotToken,
		Poller: &tb.LongPoller{Timeout: 30 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
	}
	bot = b
	b.Handle("/help", func(m *tb.Message) {
		sendMessage(b, m, constants.HelpMessage)
	})

	b.Handle("/start", func(m *tb.Message) {
		if !m.Private() {
			b.Send(m.Sender, constants.PrivateChatOnlyMessage)
			return
		}
		logic.CreateOrUpdateUser(m)
		sendMessage(b, m, fmt.Sprintf(constants.StartMessage, m.Sender.FirstName)+constants.HelpMessage)
	})

	b.Handle("/trip", func(m *tb.Message) {
		if !m.Private() {
			b.Send(m.Sender, constants.PrivateChatOnlyMessage)
			return
		}
		user := storage.GetUser(m.Sender.ID)
		if user.ID < 1 {
			logic.CreateOrUpdateUser(m)
		}

		if !logic.CanCreateNewTrip(m.Sender.ID) {
			b.Send(m.Sender, constants.LimitedCountOfTripsMessage)
			return
		}

		err = logic.CreateTrip(m, err)
		if err != nil {
			fmt.Println(err)
		}
		b.Send(m.Sender, constants.StepOneMessage)
	})

	b.Handle("/trips", func(m *tb.Message) {
		if !m.Private() {
			b.Send(m.Sender, constants.PrivateChatOnlyMessage)
			return
		}
		describedTrips := logic.ListOFTrips(m, b)
		b.Send(m.Sender, describedTrips, tb.ModeHTML)
	})

	b.Handle(tb.OnText, func(m *tb.Message) {
		if !m.Private() {
			b.Send(m.Sender, constants.PrivateChatOnlyMessage)
			return
		}
		user := storage.GetUser(m.Sender.ID)
		if user.ID < 1 {
			logic.CreateOrUpdateUser(m)
		}

		match, _ := regexp.MatchString(constants.SlashWithNumber, m.Text)
		if match {
			tripID, _ := strconv.Atoi(strings.Replace(m.Text, "/", "", 1))
			trip := storage.GetLiveTrip(m.Sender.ID, tripID)
			if trip.ID <= 0 {
				b.Send(m.Sender, constants.CanNotFinishTripMessage)
				return
			}
			logic.FinishTrip(m, tripID)
			b.Send(m.Sender, fmt.Sprintf(constants.TripFinishedTemplate, logic.TripInfo(trip)), tb.ModeHTML)
			return
		}

		trips := storage.SelectAllUncompletedTrips(m.Sender.ID)
		if trips == nil || len(trips) < 1 {
			sendMessage(b, m, constants.HelpMessage)
			return
		}
		logic.TripCreation[trips[0].Stage](b, m, &trips[0])
	})

	b.Handle(tb.OnLocation, func(m *tb.Message) {
		if !m.Private() {
			b.Send(m.Sender, constants.PrivateChatOnlyMessage)
			return
		}

		user := storage.GetUser(m.Sender.ID)
		if user.ID < 1 {
			sendMessage(b, m, constants.HelpMessage)
			return
		}
		trips := storage.SelectAllUncompletedTrips(m.Sender.ID)

		if trips == nil || len(trips) < 1 || trips[0].Stage != constants.TripLocationSecondStep {
			sendMessage(b, m, constants.HelpMessage)
			return
		}
		logic.TripCreation[constants.TripLocationSecondStep](b, m, &trips[0])
	})

	b.Handle("/finishall", func(m *tb.Message) {
		logic.FinishAllTrips(m.Sender.ID)

		b.Send(m.Sender, constants.AllTripsFinishedMessage)
	})

	b.Start()
}

func sendMessage(b *tb.Bot, m *tb.Message, helpMessageText string) (*tb.Message, error) {
	return b.Send(m.Sender, helpMessageText, tb.ModeHTML)
}
func SentMessageTo(ID int, message string) {
	bot.Send(tb.Recipient(Recipient{strconv.Itoa(ID)}), message, tb.ModeHTML)
}

type Recipient struct {
	id string
}

func (r Recipient) Recipient() string {
	return r.id
}
