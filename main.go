package main

import (
	"flag"
	"monkebot/monkebot"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// parse command-line arguments
	cfgPath := flag.String("cfg", "config.json", "path to config file")
	debug := flag.Bool("debug", false, "sets log level to debug")
	flag.Parse()

	// set up logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.DateTime,
		},
	)
	if *debug {
		log.Debug().Msg("debug mode on")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	_, err := os.Stat(*cfgPath)
	if os.IsNotExist(err) {
		log.Warn().Str("path", *cfgPath).Msg("config file does not exist, creating from template")

		var file *os.File
		file, err = os.Create(*cfgPath)
		if err != nil {
			log.Fatal().Str("path", *cfgPath).Err(err).Msg("failed to create temaplate")
		}
		defer file.Close()

		var templateJSONBytes []byte
		templateJSONBytes, err = monkebot.ConfigTemplateJSON()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to generate template")
		}

		_, err = file.Write(templateJSONBytes)
		if err != nil {
			log.Fatal().Str("path", *cfgPath).Err(err).Msg("failed to create template file")
		}

		log.Info().Str("path", *cfgPath).Msgf("template created successfully, please edit the file and run the bot again")
		os.Exit(0)
	}

	if err != nil {
		log.Fatal().Err(err)
	}

	var cfg *monkebot.Config
	cfg, err = monkebot.LoadConfigFromFile(*cfgPath)
	if err != nil {
		log.Fatal().Err(err)
	}

	var mb *monkebot.Monkebot
	mb, err = monkebot.NewMonkebot(*cfg)
	if err != nil {
		log.Fatal().Err(err)
	}

	err = mb.Connect()
	if err != nil {
		log.Fatal().Err(err)
	}
}
