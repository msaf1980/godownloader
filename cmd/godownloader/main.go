package main

import (
	"fmt"
	"os"
	"time"

	config "github.com/msaf1980/godownloader/config/godownloader"
	"github.com/msaf1980/godownloader/pkg/downloader"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	dir, logLevel, cfg, err := config.Configuration(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	var saveMode downloader.SaveMode
	err = saveMode.Set(cfg.SaveMode.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	zerolog.SetGlobalLevel(logLevel)

	d := downloader.NewDownloader(saveMode, cfg.Retry, cfg.Timeout, cfg.MaxRedirects)
	for i := range cfg.Urls {
		if !d.AddRootURL(cfg.Urls[i].URL, cfg.Urls[i].Level, cfg.Urls[i].DownLevel, cfg.Urls[i].ExtLevel) {
			log.Fatal().Str("url", cfg.Urls[i].URL).Msg("already added")
		}
	}

	switch os.Args[1] {
	case "new":
		_, err = d.NewLoad(dir, config.MAP_FILE)
		if err == nil {
			err = config.SaveConfig(dir, cfg)
		}
	case "continue":
		_, err = d.ExistingLoad(dir, config.MAP_FILE)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: '%s'\n", os.Args[1])
		os.Exit(1)
	}
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	d.Start(cfg.Parallel)
	time.Sleep(10 * time.Millisecond)
	if d.Wait() {
		log.Error().Msg("Exit with errors")
		os.Exit(1)
	} else {
		log.Info().Msg("Exit")
		os.Exit(0)
	}

}
