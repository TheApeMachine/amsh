package cmd

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/logger"
	"github.com/theapemachine/amsh/tweaker"
	"github.com/theapemachine/amsh/utils"
)

/*
Embed a mini filesystem into the binary to hold the default config file.
This will be written to the home directory of the user running the service,
which allows a developer to easily override the config file.
*/
//go:embed cfg/*
var embedded embed.FS

var (
	projectName = "amsh"
	setup       string
	cfgFile     string

	rootCmd = &cobra.Command{
		Use:   "amsh",
		Short: "A minimal shell and vim-like text editor with A.I. capabilities",
		Long:  roottxt,
	}
)

func Execute() error {
	var (
		currentDir string
		err        error
	)

	if currentDir, err = os.Getwd(); err != nil {
		fmt.Println("Error getting current working directory:", err)
		return err
	}

	if err := logger.Init(filepath.Join(currentDir, projectName+".log")); err != nil {
		fmt.Println("Error initializing logger:", err)
		return err
	}
	defer logger.Close()

	logger.Info("Logger initialized successfully")
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(
		&cfgFile, "config", "config.yml", "config file (default is $HOME/."+projectName+"/config.yml)",
	)

	rootCmd.PersistentFlags().StringVar(
		&setup, "setup", "simulation", "setup to use",
	)
}

func initConfig() {
	var err error

	if err = writeConfig(); err != nil {
		logger.Error("Error writing config: %v", err)
		log.Fatal(err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("$HOME/." + projectName)

	if err = viper.ReadInConfig(); err != nil {
		logger.Error("Failed to read config file: %v", err)
		log.Println("failed to read config file", err)
		return
	}

	tweaker.SetViper(viper.GetViper())
	logger.Info("Config initialized successfully")
}

func writeConfig() (err error) {
	var (
		home, _ = os.UserHomeDir()
		fh      fs.File
		buf     bytes.Buffer
	)

	fullPath := home + "/." + projectName + "/" + cfgFile

	if utils.CheckFileExists(fullPath) {
		return
	}

	if fh, err = embedded.Open("cfg/" + cfgFile); err != nil {
		return fmt.Errorf("failed to open embedded config file: %w", err)
	}

	defer fh.Close()

	if _, err = io.Copy(&buf, fh); err != nil {
		return fmt.Errorf("failed to read embedded config file: %w", err)
	}

	if err = os.Mkdir(home+"/."+projectName, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err = os.WriteFile(fullPath, buf.Bytes(), 0644); err != nil {
		logger.Error("Failed to write config file: %v", err)
		return fmt.Errorf("failed to write config file: %w", err)
	}

	logger.Info("Wrote config file to %s", fullPath)
	return
}

const roottxt = `amsh v0.0.1
A minimal shell and vim-like text editor written in Go, with integrated A.I. capabilities.
Different from other A.I. integrations, it uses multiple A.I. models that engage independently
in conversation with each other and the user, improving the developer experience and providing
a more human-like interaction.
`
