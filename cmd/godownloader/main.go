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
	d.AddRootURL("http://127.0.0.1", 1, 0, 0, true, nil)

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
