package logic

import (
	"bargo/storage"
	"strconv"
	"time"
	"regexp"
	"fmt"
	tb "github.com/AntoniBertel/telebot"
	"bargo/constants"
)

var TripCreation = map[int]func(*tb.Bot, *tb.Message, *storage.Trip){
	constants.TripLatencyFirstStep: func(b *tb.Bot, m *tb.Message, trip *storage.Trip) {
		n, _ := strconv.Atoi(m.Text)
		if n < 6 || n > 54 {
			b.Send(m.Sender, constants.StepOneMessage)
			return
		}
		trip.Duration = n
		trip.Stage = constants.TripLocationSecondStep
		trip.LastUpdateDate = time.Now()
		storage.UpdateUserTrip(trip)
		b.Send(m.Sender, constants.StepTwoMessage)
	},
	constants.TripLocationSecondStep: func(b *tb.Bot, m *tb.Message, trip *storage.Trip) {
		if len(m.Text) > 60 {

			b.Send(m.Sender, constants.DestinationLimitationMessage)
			return
		}
		if m.Location == nil && len(m.Text) < 1 {
			b.Send(m.Sender, constants.StepTwoMessage)
			return
		}
		if m.Location != nil {
			trip.Location = storage.Location{Longitude: float64(m.Location.Lng), Latitude: float64(m.Location.Lat)}
		} else {
			trip.TextLocation = m.Text
		}
		trip.Stage = constants.TripTrustNumberThirdStep
		trip.LastUpdateDate = time.Now()
		storage.UpdateUserTrip(trip)
		b.Send(m.Sender, constants.StepThreeMessage)
	},
	constants.TripTrustNumberThirdStep: func(b *tb.Bot, m *tb.Message, trip *storage.Trip) {
		match, _ := regexp.MatchString(constants.E164Numbers, m.Text)
		if !match {
			b.Send(m.Sender, constants.StepThreeMessage)
			return
		}
		trip.Stage = constants.TripSaveMessageFourthStep
		trip.TrustNumber = m.Text
		trip.LastUpdateDate = time.Now()
		storage.UpdateUserTrip(trip)
		b.Send(m.Sender, constants.StepFourMessage)
	},
	constants.TripSaveMessageFourthStep: func(b *tb.Bot, m *tb.Message, trip *storage.Trip) {
		if len(m.Text) < 1 || len(m.Text) > 220 {
			b.Send(m.Sender, constants.MessageLimitationMessage)
			return
		}
		trip.PanicMessage = m.Text
		trip.Stage = constants.TripCompletedFifthStep
		storage.UpdateUserTrip(trip)
		locationText := FormLocationMessage(trip.TextLocation, trip.Location)
		b.Send(m.Sender, fmt.Sprintf(constants.TripCreatedTemplate, trip.ID, trip.Duration, trip.PanicMessage, trip.TrustNumber, locationText, trip.ID), tb.ModeHTML)
	},
}

func FormLocationMessage(textMessage string, location storage.Location) (string) {
	if len(textMessage) < 1 {
		return fmt.Sprintf(constants.LocationMapTemplate, location.Longitude, location.Latitude)
	}
	return textMessage
}

func CreateTrip(m *tb.Message, err error) error {
	storage.DeleteAllUncompletedTrips(m.Sender.ID)
	trip := storage.Trip{}
	trip.USER_ID = m.Sender.ID
	trip.Stage = constants.TripLatencyFirstStep
	trip.LastUpdateDate = time.Now()
	err = storage.CreateUserTrip(&trip)
	return err
}

func ListOFTrips(m *tb.Message, b *tb.Bot) string {
	trips := storage.SelectAllCompletedTrips(m.Sender.ID)
	if len(trips) < 1 {
		b.Send(m.Sender, constants.NoSavedTripsMessage)
	}
	describedTrips := ""
	for _, trip := range trips {

		describedTrips += TripInfo(trip) + fmt.Sprintf(constants.FinishTripTemplate, trip.ID)
	}
	return describedTrips
}

func TripInfo(trip storage.Trip) string {
	locationText := FormLocationMessage(trip.TextLocation, trip.Location)

	return fmt.Sprintf(constants.TripInfoTemplate, trip.ID,
		trip.LastUpdateDate.Format(constants.DateFormat),
		trip.LastUpdateDate.Add(time.Hour * time.Duration(trip.Duration)).Format(constants.DateFormat),
		trip.TrustNumber, trip.PanicMessage, locationText)
}

func CanCreateNewTrip(USER_ID int) bool {
	trips := storage.SelectAllUnfinishedTrips(USER_ID)
	return trips == nil || len(trips) < 2
}

func ShortTripInfo(trip storage.Trip) string {
	return fmt.Sprintf(constants.ShortTripTemplate, trip.ID,
		trip.LastUpdateDate.Format(constants.DateFormat),
		trip.LastUpdateDate.Add(time.Hour * time.Duration(trip.Duration)).Format(constants.DateFormat))
}

func CreateOrUpdateUser(m *tb.Message) error {
	return storage.CreateOrUpdateUser(m.Sender)
}

func FinishTrip(m *tb.Message, tripID int) {
	storage.FinishTrip(m.Sender.ID, tripID)
}

func FinishAllTrips(USER_ID int) {
	storage.FinishAllTrips(USER_ID)
}
