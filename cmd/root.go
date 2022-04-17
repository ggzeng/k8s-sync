package cmd

import (
	"fmt"
	"os"

	log "k8sync/pkg/logger"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8sync/internal/config"
)

var (
	runMode string
	cfg     *config.Config
	Verbose bool
	daemon  bool
)

var rootCmd = &cobra.Command{
	Use:   "k8sync",
	Short: "k8sync is a tool for syncing kubernetes resources",
	Long:  `k8sync can sync kubernetes resources from one cluster to another`,
	Run: func(cmd *cobra.Command, args []string) {
		if config.GetBool("daemon") {
			log.Info("Starting k8sync in daemon mode")
			daemonStart(cmd, args)
		} else {
			cliStart(cmd, args)
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	PrintFullVersionInfo()
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&runMode, "mode", "m", "cli", "run mode with: cli, prod, dev, test")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&daemon, "daemon", "d", false, "run as daemon")
	rootCmd.PersistentFlags().BoolP("yaml", "y", false, "export source cluster yaml")
	rootCmd.PersistentFlags().StringP("src-kube-config", "", "", "source kube config file")
	rootCmd.PersistentFlags().StringP("src-namespace", "n", "", "source k8s namespace")
	rootCmd.PersistentFlags().StringArrayP("src-objects", "o", []string{"deployment", "service"}, "k8s object to sync")
	rootCmd.PersistentFlags().StringP("dst-kube-config", "c", "", "destination kube config file")
	rootCmd.PersistentFlags().StringP("dst-namespace", "", "", "destination k8s namespace")
	if err := viper.BindPFlag("app.yaml", rootCmd.PersistentFlags().Lookup("yaml")); err != nil {
		log.Fatal(err)
	}
	if err := viper.BindPFlag("src.kube-config", rootCmd.PersistentFlags().Lookup("src-kube-config")); err != nil {
		log.Fatal(err)
	}
	if err := viper.BindPFlag("src.namespace", rootCmd.PersistentFlags().Lookup("src-namespace")); err != nil {
		log.Fatal(err)
	}
	if err := viper.BindPFlag("src.objects", rootCmd.PersistentFlags().Lookup("src-objects")); err != nil {
		log.Fatal(err)
	}
	if err := viper.BindPFlag("dst.kube-config", rootCmd.PersistentFlags().Lookup("dst-kube-config")); err != nil {
		log.Fatal(err)
	}
	if err := viper.BindPFlag("dst.namespace", rootCmd.PersistentFlags().Lookup("dst-namespace")); err != nil {
		log.Fatal(err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	var err error
	if runMode == "cli" {
		viper.SetDefault("log.consoleStdout", true)
	} else {
		cfg, err = config.New(runMode)
		if err != nil {
			fmt.Printf("FATAL with mode %s: %s\n", runMode, err)
			os.Exit(1)
		}
		cfg.CheckMissingResourceEnvvars()
		viper.SetConfigFile(cfg.GetFilename())
		//viper.AddConfigPath(cfg.GetPath())
		//viper.SetConfigType(cfg.GetFileType())
		//viper.SetConfigName(cfg.GetFileBasename())
		viper.SetEnvPrefix(cfg.GetEnvPrefix())
	}
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
	log.Initialize()
}
