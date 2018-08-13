package main

import (
	"bargo/storage"
	"bargo/job"
	"bargo/bot"
)

func main() {
	storage.Connect()
	job.Schedule()
	bot.Start()
}
