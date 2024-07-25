package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/Akegarasu/blivedm-go/client"
	_ "github.com/Akegarasu/blivedm-go/utils"
)

var logger *slog.Logger
var config Config

// Liver represents a live streamer.
type Liver struct {
	Name   string `json:"name"`
	UID    int64  `json:"uid"`
	Roomid int    `json:"roomid"`
	Client *client.Client
	Writer *CsvWriter
	Ts     int64
	C      chan struct{}
}

// Config represents the configuration settings.
type Config struct {
	Cookies string `json:"cookies"`
	Nonpaid bool   `json:"nonpaid"`
	Dir     string
	Backdir string
}

// LoadLivers loads the list of livers from a JSON file.
func LoadLivers() map[string]*Liver {
	livers := map[string]*Liver{}
	byt, _ := os.ReadFile("livers.json")
	_ = json.Unmarshal([]byte(byt), &livers)
	return livers
}

// LoadConfig loads the configuration settings from a JSON file.
func LoadConfig() {
	byt, _ := os.ReadFile("config.json")
	_ = json.Unmarshal([]byte(byt), &config)
}

// makeDir creates a directory if it doesn't exist.
func makeDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = os.Mkdir(path, 0777)
	}
}

// getTime 返回当前时间的字符串表示，格式为"200601"。
func getTime() string {
	return time.Now().Format("200601")
}

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	makeDir("./data")
	makeDir("./backup")
	LoadConfig()
	config.Dir = "./data/"
	config.Backdir = "./backup/"

	var count string

	flag.StringVar(&count, "count", "0", "统计日期")
	flag.Parse()
	if count != "0" {
		slog.Info("统计日期:" + count)
		CountData(count)
		return
	}

	logs, _ := os.OpenFile("log.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	defer logs.Close()
	logger = slog.New(slog.NewTextHandler(logs, nil))
	slog.Info("started")

	livers := LoadLivers()
	if len(livers) == 0 {
		slog.Error("liver 超过 100")
		return
	}

	for _, liver := range livers {
		liver.Client = client.NewClient(liver.Roomid)
		liver.Client.SetCookie(config.Cookies)
		liver.Writer, _ = NewCsvWriter(liver.Name)
		liver.Ts = time.Now().Unix()
		go liver.Stream(ctx)
	}
	<-ctx.Done()
	slog.Debug("stopped")
}
