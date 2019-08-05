package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/NickRI/wallets-task/db"
	"github.com/NickRI/wallets-task/infrastructure/services"
	"github.com/NickRI/wallets-task/transport/restapi"
	"github.com/go-kit/kit/log"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
)

var branch, commit string

const (
	configFormat = "yaml"
	configName   = "config"
	configPath   = "config/walletsvc"
	dbPath       = "db"
	dbConfigName = "dbconf"
)

func main() {

	decimal.MarshalJSONWithoutQuotes = true

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	viper.SetConfigType(configFormat)
	viper.SetConfigName(configName)
	viper.AddConfigPath(configPath)

	if err := viper.ReadInConfig(); err != nil {
		logger.Log("error", "could not read config file", "reason", err)
		return
	}

	viper.AddConfigPath(dbPath)
	viper.SetConfigName(dbConfigName)

	if err := viper.MergeInConfig(); err != nil {
		logger.Log("error", "failed to read config file", "reason", err)
		return
	}

	dbConn, err := db.Init()
	if err != nil {
		logger.Log("error", "failed init database", "reason", err)
		return
	}

	hostAddress := fmt.Sprintf("%s:%s", viper.GetString("HOST"), viper.GetString("PORT"))

	wSvc, err := services.NewWalletService(dbConn)
	if err != nil {
		logger.Log("error", "failed init wallet service", "reason", err)
		return
	}

	routes := restapi.MakeRoutes(wSvc, logger)
	server := restapi.NewServer(hostAddress, routes)

	go server.Run()

	logger.Log("branch", branch, "commit", commit, "msg", "server is running", "host", hostAddress)

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	<-stop
	logger.Log("stop", "interrupt signal")

	server.Shutdown()
}
