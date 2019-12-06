package main

import (
	"encoding/json"
	"fmt"
	"github.com/onrik/logrus/filename"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"osm-parser/cmd"
	"strings"
)

var version string
var buildTime string

const (
	appName   = "osm-parser"
	envPrefix = "OSMP"
)

var (
	cfgFile  string
	logLevel string
	// RootCmd is the root command.
	RootCmd = &cobra.Command{
		Use:               appName,
		Short:             appName + " is parser for parsing osm data.",
		Long:              appName + " is parser for parsing osm data.",
		PersistentPreRunE: InitAll,
	}
)

// InitAll init viper & logrus.
func InitAll(cmd *cobra.Command, args []string) error {
	fmt.Println("InitAll")
	if err := InitViper(cmd, args); err != nil {
		return err
	}
	if err := InitLogrus(cmd, args); err != nil {
		return err
	}
	// Settings pretty print.
	b, err := json.MarshalIndent(viper.AllSettings(), "", " ")
	if err != nil {
		logrus.Warning(err)
	}
	logrus.Info(string(b))
	return nil
}

// InitLogrus init logrus.
func InitLogrus(cmd *cobra.Command, args []string) error {
	// Set log formatter.
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000",
		FullTimestamp:   true,
	})

	// Set log level
	logLevel := viper.GetString("LOG_LEVEL")

	switch strings.ToLower(logLevel) {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warning":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default: // default log level.
		logrus.SetLevel(logrus.DebugLevel)
	}
	logrus.Infof("logLevel: %v", logrus.GetLevel().String())

	// Add filename hook.
	// Print filename & line number as source
	logrus.AddHook(filename.NewHook())

	logrus.Info("Init Logrus success!!")
	return nil
}

// InitViper setup viper use config and load env variables.
func InitViper(cmd *cobra.Command, args []string) error {
	// Get config file if set.
	var cfgFile string
	if cfgFlag := cmd.Flags().Lookup("config"); cfgFlag != nil {
		cfgFile = cfgFlag.Value.String()
	}

	// ENV
	viper.SetDefault("config_file", "default")
	viper.AutomaticEnv()
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	logrus.Infof("config_file: %v", viper.GetString("config_file"))

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Yaml
		viper.SetConfigName(viper.GetString("config_file"))
		viper.SetConfigType("yaml")
		viper.AddConfigPath("../config")
		viper.AddConfigPath("./config")
	}

	if err := bindFlags(cmd); err != nil {
		return err
	}

	// Load yaml settings.
	if err := viper.ReadInConfig(); err == nil {
		logrus.Infof("Using config file: %v", viper.ConfigFileUsed())
	} else {
		logrus.Warningf("%v", err)
	}
	return nil
}

// Passing cmd.Commands to viper.
func bindFlags(cmd *cobra.Command) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	for _, subcmd := range cmd.Commands() {
		if err := bindFlags(subcmd); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	// config file.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is config/%s.yaml)", "default"))
	// Log level.
	RootCmd.PersistentFlags().StringVar(&logLevel, "log_level", "debug", fmt.Sprintf("Log Level (default is %s)", "DEBUG"))

	// Add cmd
	RootCmd.AddCommand(cmd.OSMParserCmd)
}

func main() {
	// Prefix print.
	// fmt.Println(logo)
	fmt.Println("Version: ", version)
	fmt.Println("BuildTime: ", buildTime)
	fmt.Println()
	if err := RootCmd.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
