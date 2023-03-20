package main

import (
	"log"

	proxyPlugin "github.com/iamelevich/pocketbase-plugin-proxy"
	"github.com/pocketbase/pocketbase"
)

func main() {
	app := pocketbase.New()

	// Setup ngrok
	proxyPlugin.MustRegister(app, &proxyPlugin.Options{
		Enabled: true,
		Url:     "http://localhost:3000",
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
