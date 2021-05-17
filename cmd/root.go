/*
SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
)

const (
	envPrefix        = "GL"
	envGardenHomeDir = envPrefix + "_HOME"
	envConfigName    = envPrefix + "_CONFIG_NAME"

	gardenHomeFolder = ".garden"
	configName       = "garden-login"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "garden-login",
	Short: "garden-login is a kubectl credential plugin for shoot cluster admin authentication",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is $HOME/%s/%s.yaml", gardenHomeFolder, configName))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		configPath := filepath.Join(home, gardenHomeFolder)

		// Search config in $HOME/.garden or in path provided with the env variable GL_HOME with name ".garden-login" (without extension) or name from env variable GL_CONFIG_NAME.
		envHomeDir, err := homedir.Expand(os.Getenv(envGardenHomeDir))
		cobra.CheckErr(err)

		viper.AddConfigPath(envHomeDir)
		viper.AddConfigPath(configPath)
		if os.Getenv(envConfigName) != "" {
			viper.SetConfigName(os.Getenv(envConfigName))
		} else {
			viper.SetConfigName(configName)
		}
	}

	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		klog.Errorf("failed to read config file: %v", err)
	}

	getClientCertificateCmd.Flags().VisitAll(func(flag *pflag.Flag) {
		viperKey := strcase.ToLowerCamel(flag.Name)

		if strings.Contains(flag.Name, "-") {
			envVarSuffix := strcase.ToScreamingSnake(flag.Name)
			envVar := fmt.Sprintf("%s_%s", envPrefix, envVarSuffix)
			if err := viper.BindEnv(viperKey, envVar); err != nil {
				klog.Warningf("Failed to bind config key %s to env variable %s: %s\n", viperKey, envVar, err.Error())
			}
		}

		viperConfigSet := viper.IsSet(viperKey)
		if !flag.Changed && viperConfigSet {
			val := viper.Get(viperKey)
			err := getClientCertificateCmd.Flags().Set(flag.Name, fmt.Sprintf("%v", val))
			if err != nil {
				klog.Warningf("Failed to set flag %s: %s\n", flag.Name, err.Error())
			}
		}
	})
}
