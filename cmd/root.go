// Copyright © 2018 Cisco Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/cisco-sso/kdk/internal/app/kdk"
	"github.com/docker/docker/client"
	"github.com/ghodss/yaml"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	KdkName string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kdk",
	Short: "Kubernetes Development Kit",
	Long: `

 _  __ ____  _  __
/ |/ //  _ \/ |/ /
|   / | | \||   / 
|   \ | |_/||   \ 
\_|\_\\____/\_|\_\
                  

A full kubernetes development environment in a container`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Fatal("Failed to execute RootCmd.")
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	kdk.Version = "0.7.0"
	rootCmd.PersistentFlags().StringVar(&KdkName, "name", "kdk", "KDK name")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func initConfig() {
	kdk.Verbose = verbose

	home, err := homedir.Dir()
	if err != nil {
		logrus.WithField("err", err).Fatal("Unable to find Home Directory")
	}

	kdk.ConfigDir = filepath.Join(home, ".kdk")
	kdk.ConfigName = "config"
	kdk.ConfigPath = filepath.Join(kdk.ConfigDir, KdkName, kdk.ConfigName+".yaml")
	kdk.KeypairDir = filepath.Join(kdk.ConfigDir, "ssh")
	kdk.PrivateKeyPath = filepath.Join(kdk.KeypairDir, "id_rsa")
	kdk.PublicKeyPath = filepath.Join(kdk.KeypairDir, "id_rsa.pub")

	if _, err := os.Stat(kdk.ConfigDir); os.IsNotExist(err) {
		err = os.Mkdir(kdk.ConfigDir, 0700)
		if err != nil {
			logrus.WithField("err", err).Fatal("Unable to create Config Directory")
		}
	}

	viper.SetConfigFile(kdk.ConfigPath)
	viper.AddConfigPath(filepath.Dir(kdk.ConfigPath))
	viper.SetConfigName(kdk.ConfigName)

	viper.SetEnvPrefix("kdk")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logrus.WithFields(logrus.Fields{"configFileUsed": viper.ConfigFileUsed(), "err": err}).Warnln("Failed to load KDK config.")
	}

	if viper.GetBool("json") {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	if _, err := os.Stat(kdk.ConfigPath); err == nil {

		// read the config.yaml file
		data, err := ioutil.ReadFile(kdk.ConfigPath)
		if err != nil {
			logrus.WithField("err", err).Fatalf("Failed to read configFile %v", kdk.ConfigPath)
		}

		err = yaml.Unmarshal(data, &kdk.KdkConfig)
		if err != nil {
			logrus.WithField("err", err).Error("Corrupted or deprecated kdk config file format")
			logrus.Fatal("Please rebuild config file with `kdk init`")
		}
	}
	kdk.Ctx = context.Background()

	kdk.DockerClient, err = client.NewEnvClient()
	if err != nil {
		logrus.WithField("err", err).Fatal("Failed to create docker client")
	}
}
