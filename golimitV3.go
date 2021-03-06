package main

import (
	"flag"
	"fmt"
	"github.com/myntra/golimit/store"
	"github.com/myntra/golimit/store/http"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type StoreConfig struct {
	ClusterName       string  `yaml:clustername,json:"clustername"`
	NodeId            string  `yaml:nodeid,json:"nodeid"`
	HostName          string  `yaml:hostname,json:"hostname"`
	TChannelPort      string  `yaml:tchannelport,json:"tchannelport"`
	Seed              string  `yaml:seednodes,json:"seednodes"`
	SyncBuffer        int     `yaml:syncbuffer,json:"syncbuffer"`
	Buckets           int     `yaml:buckets,json:"buckets"`
	StatsDEnabled     bool    `yaml:statsdenabled,json:"statsdenabled"`
	HttpPort          int     `yaml:httpport,json:"httpport"`
	UnsyncedCtrLimit  int32   `yaml:unsyncedctrlimit,json:"unsyncedctrlimit"`
	UnsyncedTimeLimit int     `yaml:unsyncedtimelimit,json:"unsyncedtimelimit"`
	StatsDHostPort    string  `yaml:statsdhostport,json:"statsdhostport"`
	StatsDSampleRate  float32 `yaml:statsdsamplerate,json:"statsdsamplerate"`
	StatsDBucket      string  `yaml:statsdbucket,json:"statsdbucket"`
	GcInterval        int     `yaml:gcinterval,json:"gcinterval"`
	ApiSecret         string  `yaml:apisecret,json:"apisecret"`
	GcGrace           int     `yaml:gcgrace,json:"gcgrace"`
}

func main() {

	configfileName := "golimit8080.yml"
	flag.StringVar(&configfileName, "config", configfileName, "Configuration File")
	loglevel := "info"
	flag.StringVar(&loglevel, "loglevel", loglevel, "Log Level")
	flag.Parse()

	log.Infof(loglevel)
	switch loglevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
	log.Infof("Log Level %+v", log.GetLevel())
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05:000"
	customFormatter.FullTimestamp = true
	log.SetFormatter(customFormatter)
	log.Info("Starting Go limiter")

	configObj := StoreConfig{}
	{
		bytes, err := ioutil.ReadFile(configfileName)
		if err != nil {
			log.Error(err)
			return
		}
		err = yaml.Unmarshal(bytes, &configObj)
		if err != nil {
			panic(err)
		}
	}
	if strings.TrimSpace(configObj.HostName) == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Errorf("Not able to resolve hostname %+v", err)
			panic(err)
		}
		configObj.HostName = hostname
	}
	log.Infof("StoreConfig: %+v", configObj)

	store := store.NewStore(
		store.WithClusterName(configObj.ClusterName),
		store.WithHttpPort(strconv.Itoa(configObj.HttpPort)),
		store.WithSeed(configObj.Seed),
		store.WithTChannelPort(configObj.TChannelPort),
	)

	http.NewGoHttpServer(configObj.HttpPort, configObj.HostName, store)

	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	sig := <-gracefulStop
	fmt.Printf("caught sig: %+v ", sig)
	store.Close()
	fmt.Println("Wait for 2 second to finish processing")
	time.Sleep(2 * time.Second)
	os.Exit(0)

}
