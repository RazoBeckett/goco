package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/razobeckett/goco/config"
	"github.com/spf13/cobra"
)

var cfg *config.Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goco",
	Short: "A conventional commit generator",
	Long:  `A CLI tool to generate conventional commit messages using Google Gemini.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		cmd.SetArgs(append([]string{"generate"}, args...))
		cmd.Execute()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.goco.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}

func initConfig() {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
}

func GetConfig() *config.Config {
	return cfg
}
