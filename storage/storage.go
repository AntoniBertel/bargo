package storage

import (
	"log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	tb "github.com/AntoniBertel/telebot"
	"fmt"
	"bargo/constants"
	"bytes"
	"encoding/binary"
	"database/sql/driver"
	"errors"
	"time"
	"database/sql"
)

var db *sqlx.DB

func Connect() {
	localDb, err := sqlx.Connect(constants.OutsideConstants.DataBaseType, constants.OutsideConstants.DataBaseConnectionString)
	db = localDb
	db.SetMaxIdleConns(constants.OutsideConstants.MaxIdleConnection)
	db.SetConnMaxLifetime(time.Duration(constants.OutsideConstants.ConnMaxLifetime) * time.Second)
	db.SetMaxOpenConns(constants.OutsideConstants.MaxOpenConnections)
	if err != nil {
		log.Fatalln(err)
	}
}

func (l Location) Value() (driver.Value, error) {
	return driver.Value(fmt.Sprintf("Point(%f %f)", l.Longitude, l.Latitude)), nil
}

func (l *Location) Scan(src interface{}) error {
	switch src.(type) {
	case []byte:
		var b = src.([]byte)
		if len(b) != 25 {
			return errors.New(fmt.Sprintf("Expected []bytes with length 25, got %d", len(b)))
		}
		var longitude float64
		var latitude float64
		buf := bytes.NewReader(b[9:17])
		err := binary.Read(buf, binary.LittleEndian, &longitude)
		if err != nil {
			return err
		}
		buf = bytes.NewReader(b[17:25])
		err = binary.Read(buf, binary.LittleEndian, &latitude)
		if err != nil {
			return err
		}
		*l = Location{longitude, latitude}
	case nil:
		*l = Location{}
	default:
		return errors.New(fmt.Sprintf("Expected []byte for Location type, got  %T", src))
	}
	return nil
}

func GetUser(ID int) (User) {
	var user = User{}
	err := db.Get(&user, `SELECT * FROM users WHERE ID = ?`, ID)
	if err != nil {
		return user
	}
	return user
}
func GetLiveTrip(USER_ID int, ID int) (Trip) {
	var trip = Trip{}
	err := db.Get(&trip, `SELECT * FROM trips WHERE ID = ? AND USER_ID = ? AND Stage < ?`, ID, USER_ID, constants.TripFinishedNinthStep)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println(err)
		return trip
	}
	return trip
}

func CreateOrUpdateUser(user *tb.User) (error) {
	tx := db.MustBegin()
	_, err := tx.NamedExec("INSERT INTO users (ID, FirstName, LastName, LanguageCode) VALUES (:ID, :FirstName, :LastName, :LanguageCode) ON DUPLICATE KEY UPDATE FirstName=:FirstName, LastName=:LastName, LanguageCode=:LanguageCode",
		User{user.ID, user.FirstName, user.LastName, user.LanguageCode})
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}
	return tx.Commit()
}

func CreateUserTrip(trip *Trip) error {
	tx := db.MustBegin()
	_, err := tx.NamedQuery(`INSERT INTO trips (USER_ID, Stage, LastUpdateDate) VALUES (:USER_ID, :Stage, NOW());`, trip)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}
	return tx.Commit()
}

func UpdateUserTrip(trip *Trip) error {
	tx := db.MustBegin()
	_, err := tx.NamedExec(`UPDATE trips SET Stage = :Stage, Duration = :Duration, PanicMessage = :PanicMessage, 
		TrustNumber = :TrustNumber, Location = ST_GeomFromText(:Location), TextLocation = :TextLocation, LastUpdateDate =NOW() WHERE ID = :ID;`,
		trip)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}
	return tx.Commit()
}

func DeleteAllUncompletedTrips(USER_ID int) error {
	tx := db.MustBegin()
	tx.MustExec(`DELETE FROM trips WHERE USER_ID = ? AND Stage < ?;`, USER_ID, constants.TripSaveMessageFourthStep)
	return tx.Commit()
}

func SelectAllUncompletedTrips(USER_ID int) ([]Trip) {
	return selectAllTripsLessThan(USER_ID, constants.TripCompletedFifthStep)
}

func SelectAllUnfinishedTrips(USER_ID int) ([]Trip) {
	return selectAllTripsLessThan(USER_ID, constants.TripFinishedNinthStep)
}

func selectAllTripsLessThan(USER_ID int, stage int) ([]Trip) {
	var trips []Trip
	err := db.Select(&trips, `SELECT * FROM trips WHERE USER_ID = ? AND Stage < ?`, USER_ID, stage)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println(err)
	}
	return trips
}
func SelectAllCompletedTrips(USER_ID int) ([]Trip) {
	var trips []Trip
	err := db.Select(&trips, `SELECT * FROM trips WHERE USER_ID = ? AND Stage >= ? AND Stage < ?`, USER_ID, constants.TripCompletedFifthStep, constants.TripFinishedNinthStep)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println(err)
	}
	return trips
}

func FinishTrip(USER_ID int, ID int) error {
	tx := db.MustBegin()
	tx.MustExec(`UPDATE trips SET Stage = ? WHERE ID = ? AND USER_ID = ?`, constants.TripFinishedNinthStep, ID, USER_ID, )
	return tx.Commit()
}

func FinishAllTrips(USER_ID int) error {
	tx := db.MustBegin()
	tx.MustExec(`UPDATE trips SET Stage = ? WHERE USER_ID = ? AND Stage >= ? AND Stage != ?`, constants.TripFinishedNinthStep, USER_ID, constants.TripCompletedFifthStep, constants.TripFinishedNinthStep)
	return tx.Commit()
}

func SelectAllExpiredTrips() ([]Trip) {
	var trips []Trip
	err := db.Select(&trips, `SELECT * FROM trips WHERE Stage = ? AND DATE_ADD(trips.LastUpdateDate, INTERVAL trips.Duration HOUR) < NOW() AND (SELECT ID FROM notifications WHERE TRIP_ID = trips.ID) IS NOT NULL`, constants.NotificationSentSixthStep)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println(err)
	}
	return trips
}

func SelectAllAlmostExpiredTrips() ([]Trip) {
	var trips []Trip
	err := db.Select(&trips, `SELECT * FROM trips WHERE Stage = ? AND DATE_ADD(trips.LastUpdateDate, INTERVAL trips.Duration -1 HOUR) < NOW() AND (SELECT ID FROM notifications WHERE TRIP_ID = trips.ID) IS NULL`, constants.TripCompletedFifthStep)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println(err)
	}
	return trips
}

func NotificationSent(TRIP_ID int) error {
	tx := db.MustBegin()
	tx.MustExec(`UPDATE trips SET Stage = ? WHERE ID = ?`, constants.NotificationSentSixthStep, TRIP_ID)
	tx.MustExec(`INSERT INTO notifications (TRIP_ID) VALUES (?)`, TRIP_ID)
	return tx.Commit()
}

func SmsSent(TRIP_ID int) error {
	tx := db.MustBegin()
	tx.MustExec(`UPDATE trips SET Stage = ? WHERE ID = ?`, constants.SmsSentSeventhStep, TRIP_ID)
	tx.MustExec(`INSERT INTO sms (TRIP_ID) VALUES (?);`, TRIP_ID)
	return tx.Commit()
}

type User struct {
	ID           int    `db:"ID"`
	FirstName    string `db:"FirstName"`
	LastName     string `db:"LastName"`
	LanguageCode string `db:"LanguageCode"`
}
type Trip struct {
	ID             int       `db:"ID"`
	USER_ID        int       `db:"USER_ID"`
	Stage          int       `db:"Stage"`
	Duration       int       `db:"Duration"`
	PanicMessage   string    `db:"PanicMessage"`
	TrustNumber    string    `db:"TrustNumber"`
	Location       Location  `db:"Location"`
	TextLocation   string    `db:"TextLocation"`
	LastUpdateDate time.Time `db:"LastUpdateDate"`
}

type Notification struct {
	ID      int `db:"ID"`
	TRIP_ID int `db:"TRIP_ID"`
}

type sms struct {
	ID      int `db:"ID"`
	TRIP_ID int `db:"TRIP_ID"`
}

type Location struct {
	Longitude float64
	Latitude  float64
}

const Schema = `
CREATE DATABASE IF NOT EXISTS bargo;
USE bargo;
ALTER DATABASE bargo CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS users(
   ID INT NOT NULL,
   FirstName VARCHAR(100) NOT NULL,
   LastName VARCHAR(100) NOT NULL,
   LanguageCode VARCHAR(40) NOT NULL,
   PRIMARY KEY ( ID )
);

CREATE TABLE IF NOT EXISTS trips(
   ID INT NOT NULL AUTO_INCREMENT,
   USER_ID INT NOT NULL,
   Stage INT DEFAULT 0,
   Duration INT DEFAULT 0,
   PanicMessage VARCHAR(360) NOT NULL DEFAULT '',
   TrustNumber VARCHAR(15) NOT NULL DEFAULT '',
   Location POINT NULL DEFAULT NULL,
   TextLocation VARCHAR(360) NOT NULL DEFAULT '',
   LastUpdateDate TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
   PRIMARY KEY ( ID )
);

CREATE TABLE IF NOT EXISTS notifications(
   ID INT NOT NULL AUTO_INCREMENT,
   TRIP_ID INT NOT NULL,
   PRIMARY KEY ( ID )
);

CREATE TABLE IF NOT EXISTS sms(
   ID INT NOT NULL AUTO_INCREMENT,
   TRIP_ID INT NOT NULL,
   PRIMARY KEY ( ID )
);
`
