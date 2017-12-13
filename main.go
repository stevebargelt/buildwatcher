package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

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

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err = viper.ReadInConfig()
	err = viper.Unmarshal(&AppConfig)
	if err != nil {
		panic(fmt.Errorf("unable to decode into struct, %v", err))
	}

	// listen for C-c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	var ciServers []server.CiServer

	var trav server.Travis
	var jenk server.Jenkins

	log.Printf("Starting up %v servers\n", len(AppConfig.Servers))
	for _, serv := range AppConfig.Servers {
		switch serv.Type {
		case "travis":
			trav.Start(serv)
			ciServers = append(ciServers, &trav)
			log.Printf("Starting Travis server %s.\n", serv.Name)
		case "jenkins":
			jenk.Start(serv)
			ciServers = append(ciServers, &jenk)
			log.Printf("Starting Jenkins server %s.\n", serv.Name)
		}
	}

	// var lights []light.Light

	// for _, l := range AppConfig.Lights {
	// 	lights = append()
	// }

	// pulls this from server config
	ticker := time.NewTicker(time.Second * time.Duration(5))
	defer ticker.Stop()

	for {
		select {
		case _ = <-ticker.C:
			results := poll(ciServers)
			for k, v := range results {
				log.Printf("Server Result [%d]: %s", k, v.Result)
				for i, j := range v.BuildResults {
					log.Printf("Build Results [%d]: %s, %s", i, j.JobName, j.Result)
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

func poll(ciservers []server.CiServer) (results []server.ServerResult) {
	c := make(chan server.ServerResult)
	go func() { c <- ciservers[0].Poll() }()
	go func() { c <- ciservers[1].Poll() }()
	timeout := time.After(2000 * time.Millisecond)
	for i := 0; i < 2; i++ { //wait for two results
		select {
		case result := <-c:
			results = append(results, result)
		case <-timeout:
			log.Println("timed out")
			return
		}
	}
	return
}

// Red    "12" //GPIO18
// Yellow "18" //GPIO24
// Green  "13" //GPIO27
// buzzer "16" //GPIO23
