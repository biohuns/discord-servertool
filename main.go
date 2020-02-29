package main

import (
	"flag"
	"log"
	"os"

	"github.com/biohuns/discord-server-tool-bot/config"
	"github.com/biohuns/discord-server-tool-bot/discord"
	"github.com/biohuns/discord-server-tool-bot/gcp"
)

var (
	stop = make(chan bool)
)

func main() {
	configPath := flag.String("config", "config.json", "config file path")
	credential := flag.String("credential", "", "credential file path")
	flag.Parse()

	if err := config.Init(*configPath); err != nil {
		log.Fatalln(err)
	}

	if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", *credential); err != nil {
		log.Fatalln(err)
	}

	if err := gcp.Init(); err != nil {
		log.Fatalln(err)
	}

	if err := discord.Init(); err != nil {
		log.Fatalln(err)
	}

	log.Println("Listening...")
	<-stop //プログラムが終了しないようロック
	return
}
