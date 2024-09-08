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
		log.Printf("config file %s does not exist, creating from template", *cfgPath)

		var file *os.File
		file, err = os.Create(*cfgPath)
		if err != nil {
			log.Fatalf("failed to create temaplate '%s' with error %v", *cfgPath, err)
			os.Exit(1)
		}
		defer file.Close()

		var templateJSONBytes []byte
		templateJSONBytes, err = monkebot.ConfigTemplateJSON()
		if err != nil {
			log.Fatalf("failed to create temaplate '%s' with error %v", *cfgPath, err)
			os.Exit(1)
		}

		_, err = file.Write(templateJSONBytes)
		if err != nil {
			log.Fatalf("failed to create temaplate '%s' with error %v", *cfgPath, err)
			os.Exit(1)
		}

		log.Printf("template %s created successfully, please edit the file and run the bot again", *cfgPath)
		os.Exit(0)
	}

	if err != nil {
		log.Fatal(err)
	}

	var cfg *monkebot.Config
	cfg, err = monkebot.LoadConfigFromFile(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	var mb *monkebot.Monkebot
	mb, err = monkebot.NewMonkebot(*cfg)
	if err != nil {
		log.Fatal(err)
	}

	err = mb.Connect()
	if err != nil {
		log.Fatal(err)
	}
}
