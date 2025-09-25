package main

import (
    "os"
    "testing"
)

func TestMainExecute(t *testing.T) {
    t.Setenv("FLIGHT_DB_HOST", "localhost")
    os.Args = []string{"flight-booking", "version"}
    main()
}

