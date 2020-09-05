package main

import (
	"os"
	"time"

	"github.com/msaf1980/godownloader/pkg/downloader"
	"github.com/rs/zerolog/log"
)

func main() {
	n := 10
	d := downloader.NewDownloader(downloader.FlatMode, 1, time.Second, 1)
	url := "http://127.0.0.1"
	if !d.AddRootURL(url, 1, 0, 0, nil) {
		log.Warn().Str("url", url).Msg("already added")
	}

	_, err := d.NewLoad("test")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	d.Start(n)
	if d.Wait() {
		log.Error().Msg("Exit with errors")
		os.Exit(1)
	} else {
		log.Info().Msg("Exit")
		os.Exit(0)
	}

}
