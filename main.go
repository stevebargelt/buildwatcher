package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	"github.com/spf13/viper"
	"github.com/stevebargelt/buildwatcher/light"
	"github.com/stevebargelt/buildwatcher/server"
)

type Config struct {
	EnableGPIO bool   `yaml:"enable_gpio"`
	Database   string `yaml:"database"`
	HighRelay  bool   `json:"highrelay"`
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

	// create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// listen for C-c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// defer func() {
	// 	signal.Stop(c)
	// 	cancel()
	// }()

	// go func() {
	// 	select {
	// 	case <-c:
	// 		cancel()
	// 	case <-ctx.Done():
	// 		err := ctx.Err()
	// 		log.Println(ctx, err.Error())
	// 	}
	// }()

	//var lights []light.Light
	var trav server.Travis
	var jenk server.Jenkins

	log.Printf("Starting up %v servers\n", len(AppConfig.Servers))
	for _, serv := range AppConfig.Servers {
		switch serv.Type {
		case "travis":
			go trav.Start(ctx, serv)
			log.Printf("Starting Travis server %s.\n", serv.Name)
		case "jenkins":
			go jenk.Start(ctx, serv)
			log.Printf("Starting Travis server %s.\n", serv.Name)
		}
	}

	// // Initialize Travis
	// go trav.Start(AppConfig.Servers[1])
	// log.Println("Starting Travis polling")

	// // Initialize Jenkins
	// go jenk.Start(AppConfig.Servers[0])
	// log.Println("Starting Jenkins polling")

	// <-c
	// fmt.Println("main: received C-c - shutting down")

	// // tell the goroutines to stop
	// fmt.Println("main: telling goroutines to stop")
	// cancel()

	// // and wait for them both to reply back
	// wg.Wait()
	// fmt.Println("main: all goroutines have told us they've finished")

	// ch := make(chan os.Signal, 1)
	// signal.Notify(ch, os.Interrupt, syscall.SIGUSR2)
	for {
		select {
		case <-c:
			cancel()
			log.Println("case CTRL-C was detected... cancel called")
		case <-ctx.Done():
			err := ctx.Err()
			log.Println("HERE:", ctx, err.Error())
			return
		}
	}

}

// Red    "12" //GPIO18
// Yellow "18" //GPIO24
// Green  "13" //GPIO27
// buzzer "16" //GPIO23
