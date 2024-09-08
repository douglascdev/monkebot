package main

import (
	"flag"
	"log"
	"monkebot/monkebot"
	"os"
)

func main() {
	cfgPath := flag.String("cfg", "config.json", "path to config file")
	flag.Parse()

	_, err := os.Stat(*cfgPath)
	if os.IsNotExist(err) {
		log.Fatalf("config file %s does not exist", *cfgPath)
	}

	if err != nil {
		log.Fatal(err)
	}

	cfg, err := monkebot.LoadConfigFromFile(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	mb, err := monkebot.NewMonkebot(*cfg)
	if err != nil {
		log.Fatal(err)
	}

	err = mb.Connect()
	if err != nil {
		log.Fatal(err)
	}
}
