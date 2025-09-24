package main

import (
    "log"

    "github.com/ambiyansyah-risyal/flight-booking/internal/adapter/cli"
)

func main() {
    if err := cli.Execute(); err != nil {
        log.Fatal(err)
    }
}

