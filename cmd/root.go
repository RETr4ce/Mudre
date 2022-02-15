package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

func init() {
	cobra.OnInitialize(config)

	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("[*] INIT: ", err)
	}

	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

var RootCmd = &cobra.Command{
	Use:     "mudre",
	Version: "0.1.0",
	Long:    "Collecting information and posting it on Discord",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Use: %s --help\n", cmd.Use)
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(0)
	}
}

func config() {
	viper.SetConfigName("mudre")
	viper.AddConfigPath("./config/")
	viper.AddConfigPath("$HOME/.mudre")
	viper.AddConfigPath(".")
	viper.AddConfigPath("toml")

	err := viper.ReadInConfig()
	if err != nil {
		ErrorLogger.Println("[*] ", err)
		os.Exit(0)
	}
	viper.AutomaticEnv()
}
