package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danesparza/appupgrade/api"
	"github.com/danesparza/appupgrade/system"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Long:  `The start command starts the app upgrade server`,
	Run:   start,
}

func start(cmd *cobra.Command, args []string) {
	//	If we have a config file, report it:
	if viper.ConfigFileUsed() != "" {
		log.Debugf("Using config file: %s", viper.ConfigFileUsed())
	} else {
		log.Debug("No config file found.")
	}

	//	Get the list of packages (and their github urls)
	monitorPackages := viper.GetStringMap("packages")

	//	Emit what we know:
	log.WithFields(log.Fields{
		"Monitor packages": monitorPackages,
	}).Info("Starting up")

	//	Create an api service object
	apiService := api.Service{
		StartTime: time.Now(),
	}

	//	Trap program exit appropriately
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go handleSignals(ctx, sigs, cancel)

	//	Log that the system has started:
	log.Info("System started")

	//	Create a router and setup our REST and UI endpoints...
	restRouter := mux.NewRouter()

	//	VERSION ROUTES
	restRouter.HandleFunc("/v1/versioninfo/{package}", apiService.GetVersionInfoForPackage).Methods("GET") // Get version data

	//	Format the bound interface:
	formattedServerInterface := viper.GetString("server.bind")
	if formattedServerInterface == "" {
		outboundIP, err := system.GetOutboundIP()
		if err == nil {
			formattedServerInterface = outboundIP.String()
		}
	}

	//	Start the service and display how to access it
	go func() {
		formattedServiceURL := fmt.Sprintf("http://%s:%s/v1/swagger/", formattedServerInterface, viper.GetString("server.port"))
		log.WithFields(log.Fields{
			"url": formattedServiceURL,
		}).Info("Started REST service")
		log.Printf("[ERROR] %v\n", http.ListenAndServe(viper.GetString("server.bind")+":"+viper.GetString("server.port"), restRouter))
	}()

	//	Wait for our signal and shutdown gracefully
	<-ctx.Done()
}

func handleSignals(ctx context.Context, sigs <-chan os.Signal, cancel context.CancelFunc) {
	select {
	case <-ctx.Done():
	case sig := <-sigs:
		switch sig {
		case os.Interrupt:
			log.WithFields(log.Fields{
				"signal": "SIGINT",
			}).Info("Shutting down")
		case syscall.SIGTERM:
			log.WithFields(log.Fields{
				"signal": "SIGTERM",
			}).Info("Shutting down")
		}

		cancel()
		os.Exit(0)
	}
}

func init() {
	rootCmd.AddCommand(startCmd)
}
