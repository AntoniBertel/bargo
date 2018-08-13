package job

import (
	"fmt"
	"log"
	"bargo/storage"
	"bargo/logic"
	"bargo/bot"
	"github.com/sfreiberg/gotwilio"
	"bargo/constants"
	"github.com/robfig/cron"
)

var twilio = gotwilio.NewTwilioClient(constants.OutsideConstants.TwilioKey, constants.OutsideConstants.TwilioAuthToken)

func Schedule() {
	c := cron.New()
	err := c.AddFunc("@every 30s", notifyInChatIfNeeded)
	if err != nil {
		log.Fatalln(err)
	}
	err = c.AddFunc("@every 60s", sendSmsIfNeeded)
	if err != nil {
		log.Fatalln(err)
	}
	c.Start()
}

func sendSmsIfNeeded() {
	trips := storage.SelectAllExpiredTrips()
	for _, trip := range trips {
		user := storage.GetUser(trip.USER_ID)
		locationText := logic.FormLocationMessage(trip.TextLocation, trip.Location)
		err, exception := sendSms(trip.TrustNumber,
			fmt.Sprintf(constants.SMSTemplate, user.FirstName, user.LastName, trip.PanicMessage, locationText))
		if err != nil {
			log.Println(err)
			return
		}
		if err != nil {
			log.Println(exception)
			return
		}
		sendSmsSentInformation(trip)
		storage.SmsSent(trip.ID)

	}
}

func notifyInChatIfNeeded() {
	trips := storage.SelectAllAlmostExpiredTrips()
	for _, trip := range trips {
		sendNotification(trip)
		storage.NotificationSent(trip.ID)
	}
}

func sendSms(number string, text string) (error, *gotwilio.Exception) {
	_, exception, err := twilio.SendSMS(constants.OutsideConstants.TwilioFromNumber, number, text, "", "")
	if exception != nil || err != nil {
		return nil, exception
	}
	return nil, nil
}

func sendNotification(trip storage.Trip) {
	bot.SentMessageTo(trip.USER_ID, fmt.Sprintf(constants.TripAlmostFinishedTemplate, logic.ShortTripInfo(trip), trip.ID))
}

func sendSmsSentInformation(trip storage.Trip) {
	bot.SentMessageTo(trip.USER_ID, fmt.Sprintf(constants.SmsSentTemplate, trip.PanicMessage, trip.TrustNumber, trip.ID))
}
