package config

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/msaf1980/godownloader/pkg/downloader"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
)

const (
	CONFIG_FILE = "godownloader.yml"
	MAP_FILE    = "godownloader.map"
)

type URL struct {
	URL       string `yaml:"url"`
	Level     int32  `yaml:"level"`
	DownLevel int32  `yaml:"down_level"`
	ExtLevel  int32  `yaml:"ext_level"`
}

type URLslice []URL

func (u *URLslice) Set(value string) (err error) {
	s := strings.Split(value, " ")
	if len(s) != 4 {
		return fmt.Errorf("url must have format 'url1 LEVEL DOWN_LEVEL EXT_LEVEL': '%s'", value)
	}
	url := URL{URL: s[0]}
	var tmp int64

	tmp, err = strconv.ParseInt(s[1], 10, 32)
	if err != nil {
		return fmt.Errorf("url level must be a number: '%s'", s[1])
	}
	url.Level = int32(tmp)

	tmp, err = strconv.ParseInt(s[2], 10, 32)
	if err != nil {
		return fmt.Errorf("url downLevel must be a number: '%s'", s[2])
	}
	url.DownLevel = int32(tmp)

	tmp, err = strconv.ParseInt(s[3], 10, 32)
	if err != nil {
		return fmt.Errorf("url extLevel must be a number: '%s'", s[3])
	}
	url.ExtLevel = int32(tmp)

	*u = append(*u, url)
	return nil
}

func (u *URLslice) String() string {
	return fmt.Sprintf("%+v", *u)
}

type SaveModeStr string

func (s *SaveModeStr) Set(value string) (err error) {
	m := downloader.FlatMode
	err = m.Set(value)
	if err == nil {
		*s = SaveModeStr(value)
	}
	return
}

func (s *SaveModeStr) String() string {
	return string(*s)
}

type LogLevel string

func (l *LogLevel) Set(value string) error {
	level := strings.ToLower(value)
	switch level {
	case "info", "warn", "debug":
		*l = LogLevel(level)
		return nil
	default:
		return fmt.Errorf("incorrect loglevel %s", value)
	}
}

func (l *LogLevel) String() string {
	return string(*l)
}

func (l *LogLevel) Level() zerolog.Level {
	switch *l {
	case "info":
		return zerolog.InfoLevel
	case "debug":
		return zerolog.DebugLevel
	default:
		return zerolog.WarnLevel
	}
}

type Config struct {
	Urls         URLslice      `yaml:"urls"`
	Retry        int           `yaml:"retry"`
	MaxRedirects int           `yaml:"max_redirects"`
	Timeout      time.Duration `yaml:"timeout"`
	SaveMode     SaveModeStr   `yaml:"save_mode"`
	Parallel     int
}

func defaultConfig() *Config {
	cfg := &Config{
		Urls:         make([]URL, 0),
		Retry:        3,
		MaxRedirects: 0,
		Timeout:      1 * time.Second,
		SaveMode:     "flat",
		Parallel:     1,
	}

	return cfg
}

func LoadConfig(dir string, cfg *Config) error {
	yml, err := ioutil.ReadFile(dir + "/" + CONFIG_FILE)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yml, cfg)
	if err != nil {
		return err
	}
	return nil
}

func SaveConfig(dir string, cfg *Config) error {
	yml, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dir+"/"+CONFIG_FILE, yml, 0o644)
}

// LoadConfig load config file/parse cmd args
func Configuration(args []string) (string, zerolog.Level, *Config, error) {
	cfg := defaultConfig()

	showHelp := false
	var dir string
	logLevel := LogLevel("warn")

	flagNew := flag.NewFlagSet("new", flag.ContinueOnError)
	flagNew.StringVar(&dir, "dir", "", "out dir")
	flagNew.IntVar(&cfg.Parallel, "parallel", 1, "parallel")
	flagNew.IntVar(&cfg.Retry, "retry", 1, "retry")
	flagNew.IntVar(&cfg.MaxRedirects, "redirects", 0, "max redirects")
	flagNew.DurationVar(&cfg.Timeout, "timeout", 0, "timeout")
	flagNew.Var(&cfg.SaveMode, "save", "save mode [ flat | flat_dir | site_dir | dir ]")
	flagNew.Var(&logLevel, "loglevel", "loglevel [debug | info | warn]")
	flagNew.BoolVar(&showHelp, "help", false, "help")
	helpNew := func() {
		fmt.Fprintf(os.Stderr, "\n%s new OPTIONS 'url1 LEVEL DOWN_LEVEL EXT_LEVEL' ..\n", args[0])
		flagNew.Usage()
	}

	flagCont := flag.NewFlagSet("continue", flag.ContinueOnError)
	flagCont.StringVar(&dir, "dir", "", "out dir")
	flagCont.IntVar(&cfg.Parallel, "parallel", 1, "parallel")
	flagCont.Var(&logLevel, "loglevel", "loglevel [debug | info | warn]")
	flagCont.BoolVar(&showHelp, "help", false, "help")
	helpCont := func() {
		fmt.Fprintf(os.Stderr, "\n%s continue OPTIONS\n", args[0])
		flagCont.Usage()
	}

	helpAll := func() {
		fmt.Fprintf(os.Stderr, "%s: mirror of http sites\n", args[0])
		helpNew()
		helpCont()
	}

	if len(args) > 1 {
		switch os.Args[1] {
		case "new":
			err := flagNew.Parse(args[2:])
			if err == nil && showHelp {
				helpNew()
			}
			if err != nil || showHelp {
				os.Exit(1)
			}
			for _, value := range flagNew.Args() {
				if err = cfg.Urls.Set(value); err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", err.Error())
					os.Exit(1)
				}
			}
			if len(dir) == 0 {
				return dir, logLevel.Level(), nil, fmt.Errorf("configuration: dir not set")
			}
		case "continue":
			err := flagCont.Parse(args[2:])
			if err == nil && showHelp {
				helpCont()
			}
			if err != nil || showHelp {
				os.Exit(1)
			}
			f := flagCont.Args()
			if len(f) > 0 {
				fmt.Fprintf(os.Stderr, "non-flag arguments:\n")
				for _, value := range f {
					fmt.Fprintf(os.Stderr, "  '%s'\n", value)
				}
				helpCont()
				os.Exit(1)
			}
			if len(dir) == 0 {
				return dir, logLevel.Level(), nil, fmt.Errorf("configuration: dir not set")
			}
			err = LoadConfig(dir, cfg)
			if err != nil || showHelp {
				return dir, logLevel.Level(), nil, err
			}
		default:
			if args[1] != "-h" && args[1] != "-help" {
				fmt.Fprintf(os.Stderr, "unknown command '%s'\n\n", args[1])
			}
			showHelp = true
		}
	} else {
		showHelp = true
	}

	if showHelp {
		helpAll()
		os.Exit(1)
	}

	if len(cfg.Urls) == 0 {
		return dir, logLevel.Level(), cfg, fmt.Errorf("configuration: urls empthy")
	}
	if cfg.Retry < 1 {
		return dir, logLevel.Level(), cfg, fmt.Errorf("configuration: retry < 1")
	}
	if cfg.MaxRedirects < 0 {
		cfg.MaxRedirects = 0
	}
	if cfg.Timeout < 0 {
		cfg.Timeout = 0
	}

	return dir, logLevel.Level(), cfg, nil
}
