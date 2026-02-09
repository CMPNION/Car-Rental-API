package main

import (
	"fmt"

	"github.com/CMPNION/Car-Rental-API.git/internal/infra/database"
	"github.com/CMPNION/Car-Rental-API.git/internal/interface/http/server"
)


func main(){
    fmt.Println("Car Rental System API")

    db := database.InitDB("car_rental.db")
    app := server.GetNewServer(":4000", db)


    if err := app.Start(); err != nil {
        panic("Failed to start Server: " + err.Error())
    }
}
