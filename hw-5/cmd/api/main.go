package main

import (
	"log"
	"net/http"
	"tfs-go-hw/hw-5/internals/handlers"
	"tfs-go-hw/hw-5/internals/repository"
	"tfs-go-hw/hw-5/internals/services"
)

func main() {
	database := repository.NewDatabase()
	service := services.NewMessenger(database)
	handler := handlers.NewMessenger(service)

	log.Fatal(http.ListenAndServe(":5000", handler.Router()))
}
