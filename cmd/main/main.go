package main

import (
	"fmt"

	"github.com/CMPNION/Car-Rental-API.git/internal/infrastructure/database"
	"github.com/CMPNION/Car-Rental-API.git/internal/infrastructure/server"
)


func main(){
    fmt.Println("Car Rental System API")

    db := database.InitDB("../../car_rental.db")
    app := server.GetNewServer(":8080", db)


    if err := app.Start(); err != nil {
        panic("Failed to start Server: " + err.Error())
    }



}
