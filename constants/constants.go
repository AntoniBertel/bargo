package constants

import (
	tb "github.com/AntoniBertel/telebot"
	"github.com/kelseyhightower/envconfig"
	"log"
)

func init() {
	err := envconfig.Process("bargo", &OutsideConstants)
	if err != nil {
		log.Fatal(err.Error())
	}
}

var OutsideConstants Configuration

type Configuration struct {
	TelegramBotToken         string `default:"000"`
	TwilioKey                string `default:"000"`
	TwilioFromNumber         string `default:"000"`
	TwilioAuthToken          string `default:"000"`
	DataBaseType             string `default:"mysql"`
	DataBaseConnectionString string `default:"000:000@/000?parseTime=true"`
	MaxIdleConnection        int    `default:"2"`
	ConnMaxLifetime          int    `default:"30"`
	MaxOpenConnections       int    `default:"10"`
}

var CreateTripButton = tb.InlineButton{
	Unique: "create_trip",
	Text:   "Create trip",
}
var ListOfTripsButton = tb.InlineButton{
	Unique: "list_of_trips",
	Text:   "List of trips",
}
var FinishAllTripsButton = tb.InlineButton{
	Unique: "finish_all_trips",
	Text:   "Finish all trips",
}

var InlineKeys = [][]tb.InlineButton{
	{CreateTripButton}, {ListOfTripsButton}, {FinishAllTripsButton},
}

const StartMessage = "Hello %s, I'm Bargo and I'll automatically notify your friends if you get lost. \n" +
	"I'll ask you to set a trip duration, trust person's number, anticipated location and warning message. " +
	"I'll send SMS to your friend if you not finish your trip in time.\n"
const HelpMessage = "/trip to tell me about your future trip. \n" +
	"/trips to see and manage all live trips.\n" +
	"/finishall to finish all live trips."
const PrivateChatOnlyMessage = "I can talk only in private."
const LimitedCountOfTripsMessage = "I can't create another one yet, please finish one of the live /trips"
const CanNotFinishTripMessage = "I can't finish this trip."

const ResetMessage = "\n/trip to reset."
const StepOneMessage = "How many hours will trip take? The minimum is 6 hours, the maximum is 54 hours"
const StepTwoMessage = "What's your destination? Say me in text, or send a location. " + ResetMessage
const StepThreeMessage = "What's your trust person number? Format +14155552671. " + ResetMessage
const StepFourMessage = "What's the safe message? I'll deliver it if you don't finish your trip in time." + ResetMessage
const DestinationLimitationMessage = "I expect the destination to be less than 60 symbols"
const MessageLimitationMessage = "I expect the message to be less than 220 symbols."
const AllTripsFinishedMessage = "I finished all your trips."
const NoSavedTripsMessage = "I can't find your trips, /trip to create."

const TripCreatedTemplate = `Congratulations! I remembered the trip <b>#%d</b> duration <b>%d</b> hours. If you will not finish trip in time, I'll deliver the message <b>%s</b> to the number <b>%s</b>, with your location <b>%s</b>
I can finish this trip /%d, or show all live /trips`
const TripFinishedTemplate = "Congratulations! I finished %s \n/trip to create new"
const SMSTemplate = "%s %s notifies you: '%s', Location:%s"
const TripAlmostFinishedTemplate = "Attention! %s is nearing completion. \nFinish trip /%d"
const SmsSentTemplate = "Attention! I delivered <b>%s</b> to <b>%s</b>. \nFinish trip /%d"
const LocationMapTemplate = "https://maps.google.com/maps?q=loc:%f,%f"
const TripInfoTemplate = `Trip <b>#%d</b>, started at <b>%s</b> 
Ends <b>%s</b>
Trust number <b>%s</b>
Message <b>%s</b>
Location <b>%s</b>`
const ShortTripTemplate = `Trip <b>#%d</b>, started at <b>%s</b> 
Ends <b>%s</b>`
const FinishTripTemplate = "\nFinish trip /%d \n"
const E164Numbers = "^\\++?[0-9]\\d{6,14}$"
const SlashWithNumber = "\\/[0-9]+"
const DateFormat = "2006-01-02 15:04:05 Monday"

const (
	TripLatencyFirstStep      = 1
	TripLocationSecondStep    = 2
	TripTrustNumberThirdStep  = 3
	TripSaveMessageFourthStep = 4
	TripCompletedFifthStep    = 5
	NotificationSentSixthStep = 6
	SmsSentSeventhStep        = 7
	TripFinishedNinthStep     = 9
)
