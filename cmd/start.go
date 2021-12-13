package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/danesparza/appupgrade/api"
	_ "github.com/danesparza/appupgrade/docs" // swagger docs location
	"github.com/danesparza/appupgrade/system"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	httpSwagger "github.com/swaggo/http-swagger"
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

	//	PACKAGE ROUTES
	restRouter.HandleFunc("/v1/package/{package}/info", apiService.GetVersionInfoForPackage).Methods("GET")                     // Get version data
	restRouter.HandleFunc("/v1/package/{package}/updatetoversion/{version}", apiService.UpdatePackageToVersion).Methods("POST") // Update app to the specified version

	//	SWAGGER ROUTES
	restRouter.PathPrefix("/v1/swagger").Handler(httpSwagger.WrapHandler)

	// Setup CORS
	restCorsRouter := cors.New(cors.Options{
		AllowedOrigins:   strings.Split(viper.GetString("server.allowed-origins"), ","),
		AllowCredentials: true,
	}).Handler(restRouter)

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
		log.Printf("[ERROR] %v\n", http.ListenAndServe(viper.GetString("server.bind")+":"+viper.GetString("server.port"), restCorsRouter))
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

		//	Clean up the tmp directory area where we store .deb files?
		//	It sounds like /tmp is cleared on reboot, and files are automatically removed after 10 days.
		//	So I'm not sure we need to do anything -- we can just let the OS take care of it.

		os.Exit(0)
	}
}

func init() {
	rootCmd.AddCommand(startCmd)
}
