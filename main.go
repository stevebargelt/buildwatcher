package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	fsnotify "github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"

	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	"github.com/spf13/viper"
	"github.com/stevebargelt/buildwatcher/light"
	"github.com/stevebargelt/buildwatcher/server"
)

type Config struct {
	EnableGPIO bool   `yaml:"enable_gpio"`
	Database   string `yaml:"database"`
	HighRelay  bool   `json:"highrelay"`
	Pollrate   int    `json:"pollrate"`
	Lights     []light.Light
	Servers    []server.Server
}

// Version is the version of the app
var Version string

// AppConfig is the top level configuration for the entire app
var AppConfig *Config

var (
	ciServers []server.CiServer
	trav      server.Travis
	jenk      server.Jenkins
)

func main() {

	//create your file with desired read/write permissions
	f, err := os.OpenFile("buildwatcher.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()
	log.SetOutput(f)

	//configFile := flag.String("config", "", "Build Watcher configuration file path")
	version := flag.Bool("version", false, "Print version information")
	flag.Usage = func() {
		text := `
    Usage: buildwatcher [OPTIONS]

    Options:

      -config string
          Configuration file path
      -version
			    Print version information
    `
		fmt.Println(strings.TrimSpace(text))
	}
	flag.Parse()
	if *version {
		fmt.Println(Version)
		return
	}

	setConfig()

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Config file changed:", e.Name)
		setConfig()
		startServers()
	})

	// listen for C-c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	startServers()

	// var lights []light.Light

	// for _, l := range AppConfig.Lights {
	// 	lights = append()
	// }

	ticker := time.NewTicker(time.Second * time.Duration(AppConfig.Pollrate))
	defer ticker.Stop()

	for {
		select {
		case _ = <-ticker.C:
			poll(ciServers)
			for k, ciserver := range ciServers {
				switch v := ciserver.(type) {
				case *server.Jenkins:
					log.Printf("Jenkins: %s", v.Name)
				case *server.Travis:
					log.Printf("Travis: %s", v.Name)
				default:
					log.Fatalf("FATAL: I don't know about type %T of ciservers!\n", v)
				}
				// if jenkinsCIserver, ok := ciserver.(*server.Jenkins); ok {
				// 	log.Printf(jenkinsCIserver.Name)
				// }
				log.Printf("Server Result [%d]: %s", k, ciserver.Status())
				for i, j := range ciserver.JobStatus() {
					log.Printf("Build Results [%d]: %s", i, j)
				}
			}
		case s := <-c:
			switch s {
			case os.Interrupt:
				log.Println("CTRL-C was detected... cancel called")
				return
				// case syscall.SIGUSR2:
				// 	c.DumpTelemetry()
			}
		}
	}

}

func startServers() {

	log.Printf("Starting up %v servers\n", len(AppConfig.Servers))
	ciServers = nil
	for _, serv := range AppConfig.Servers {
		switch serv.Type {
		case "travis":
			trav.Start(serv)
			ciServers = append(ciServers, &trav)
			log.Printf("Starting a Travis server %s.\n", serv.Name)
		case "jenkins":
			jenk.Start(serv)
			ciServers = append(ciServers, &jenk)
			log.Printf("Starting a Jenkins server %s.\n", serv.Name)
		}
	}
}

func poll(ciservers []server.CiServer) {
	c := make(chan bool)
	go func() { c <- ciservers[0].Poll() }()
	go func() { c <- ciservers[1].Poll() }()
	timeout := time.After(2000 * time.Millisecond)
	for i := 0; i < 2; i++ { //wait for two results
		select {
		case _ = <-c:

		case <-timeout:
			log.Println("timed out")
			return
		}
	}
	return
}

func setConfig() {

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("unable to read config, %v", err))
	}
	err = viper.Unmarshal(&AppConfig)
	if err != nil {
		panic(fmt.Errorf("unable to decode into struct, %v", err))
	}
}

// Red    "12" //GPIO18
// Yellow "18" //GPIO24
// Green  "13" //GPIO27
// buzzer "16" //GPIO23
