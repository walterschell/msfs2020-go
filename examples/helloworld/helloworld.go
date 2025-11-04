package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/walterschell/msfs2020-go/package/simconnect"
)

type PositionReport struct {
	Latitude  float64 `name:"PLANE LATITUDE" units:"degrees"`
	Longitude float64 `name:"PLANE LONGITUDE" units:"degrees"`
	Altitude  float64 `name:"PLANE ALTITUDE" units:"feet"`
	Heading   float64 `name:"PLANE HEADING DEGREES MAGNETIC" units:"degrees"`
	Speed     float64 `name:"AIRSPEED TRUE" units:"knots"`
	AGL       float64 `name:"PLANE ALT ABOVE GROUND" units:"feet"`
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		//	Level: slog.LevelDebug, // Set log level to Debug
	})))
	port := flag.String("port", "500", "SimConnect server port (default 500)")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [host] [-port <port>]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	host := args[0]

	endpoint := fmt.Sprintf("%s:%s", host, *port)
	conn, err := simconnect.Connect(endpoint)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to SimConnect server: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing SimConnect connection: %v\n", err)
		}
	}()
	fmt.Printf("Connected to SimConnect server at %s\n", endpoint)

	var i = PositionReport{}
	if _, err := conn.StreamDataOnSimObject(i, 0, 5, context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error streaming data on SimObject: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Press Enter to Teleport to (Lat: 00.0, Lon: 0.0, Alt: 1000.0)\n")
	fmt.Scanln() // Wait for Enter key press

	// teleportRequest := simconnect.TeleportRequestWithAltitude{0.0, 0.0, 30000.0}
	// if err := conn.SetDataOnSimObject(0, teleportRequest); err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error setting data on SimObject: %v\n", err)
	// 	os.Exit(1)
	// }

	for _ = range 1 {
		coldStartRequest := simconnect.NewColdStartTeleportRequest(0.0, 0.0)
		if err := conn.SetDataOnSimObject(0, coldStartRequest); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting data on SimObject: %v\n", err)
			os.Exit(1)
		}
	}

	// if err := conn.Teleport(0.0, 0.0, simconnect.WithAltitude(1000.0)); err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error teleporting: %v\n", err)
	// 	os.Exit(1)
	// }

	// Create a channel to receive OS signals.
	sigChan := make(chan os.Signal, 1)

	// Notify the sigChan when SIGINT (Ctrl+C) or SIGTERM is received.
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Println("Program running. Press Ctrl+C to exit gracefully.")

	// Block until a signal is received on sigChan.
	sig := <-sigChan

	fmt.Printf("Received signal: %s. Performing cleanup...\n", sig)

	// Add your cleanup logic here.
	// For example, closing database connections, flushing logs, etc.
	fmt.Println("Cleanup complete. Exiting.")

}
