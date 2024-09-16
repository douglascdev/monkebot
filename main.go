package main

import (
	"bytes"
	"flag"
	"monkebot/config"
	"monkebot/database"
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
		templateJSONBytes, err = config.ConfigTemplateJSON()
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
		log.Fatal().Err(err).Msg("failed to stat config file")
	}

	var (
		cfg  *config.Config
		data []byte
	)
	data, err = os.ReadFile(*cfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read config file")
	}
	cfg, err = config.LoadConfig(data)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config file")
	}

	reader := new(bytes.Buffer)
	reader.Write(data)
	writer, err := os.OpenFile(*cfgPath, os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open config file for writing")
	}
	db, err := database.InitDB(cfg.DBConfig.Driver, cfg.DBConfig.DataSourceName, reader, writer)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize database")
	}

	var mb *monkebot.Monkebot
	mb, err = monkebot.NewMonkebot(*cfg, db)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize monkebot")
	}

	err = mb.Connect()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to Twitch")
	}
}
