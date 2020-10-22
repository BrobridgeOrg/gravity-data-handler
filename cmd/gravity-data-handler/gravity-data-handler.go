package main

import (
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	//	_ "go.uber.org/automaxprocs"

	app "github.com/BrobridgeOrg/gravity-data-handler/pkg/app/instance"
)

func init() {

	// From the environment
	viper.SetEnvPrefix("GRAVITY_DATA_HANDLER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// From config file
	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("./configs")

	if err := viper.ReadInConfig(); err != nil {
		log.Warn("No configuration file was loaded")
	}

	runtime.GOMAXPROCS(16)
	/*
		go func() {

			defer os.Exit(0)

			f, err := os.Create("trace.out")
			//f, err := os.Create("cpu-profile.prof")
			if err != nil {
				log.Fatal(err)
			}

			trace.Start(f)
			defer trace.Stop()

			//		pprof.StartCPUProfile(f)
			//		defer pprof.StopCPUProfile()

			sig := make(chan os.Signal, 1)
			signal.Notify(sig, os.Interrupt, os.Kill)
			<-sig
		}()
	*/
}

func main() {

	// Initializing application
	a := app.NewAppInstance()

	err := a.Init()
	if err != nil {
		log.Fatal(err)
		return
	}

	// Starting application
	err = a.Run()
	if err != nil {
		log.Fatal(err)
		return
	}
}
