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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/buffer"
	"github.com/theapemachine/amsh/chat"
	"github.com/theapemachine/amsh/editor"
	"github.com/theapemachine/amsh/filebrowser"
	"github.com/theapemachine/amsh/header"
	"github.com/theapemachine/amsh/logger"
	"github.com/theapemachine/amsh/splash"
	"github.com/theapemachine/amsh/statusbar"
	"github.com/theapemachine/amsh/ui"
	"github.com/theapemachine/amsh/utils"
	"golang.org/x/term"
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
	cfgFile     string
	path        string

	rootCmd = &cobra.Command{
		Use:   "amsh",
		Short: "A minimal shell and vim-like text editor with A.I. capabilities",
		Long:  roottxt,
		RunE: func(cmd *cobra.Command, args []string) error {
			width, height, _ := term.GetSize(int(os.Stdout.Fd()))

			buf := buffer.New(path, width, height)
			buf.RegisterComponents("splash", splash.New(width, height))
			buf.RegisterComponents("header", header.New(width, height))
			buf.RegisterComponents("filebrowser", filebrowser.New(width, height))
			buf.RegisterComponents("editor", editor.New(width, height))
			buf.RegisterComponents("statusbar", statusbar.New(width))
			buf.RegisterComponents("chat", chat.New(width, height))

			prog := tea.NewProgram(
				buf,
				tea.WithAltScreen(),
			)

			if _, err := prog.Run(); err != nil {
				fmt.Println("Error while running program:", err)
				os.Exit(1)
			}

			return nil
		},
	}
)

func Execute() {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current working directory:", err)
		os.Exit(1)
	}

	if err := logger.Init(filepath.Join(currentDir, projectName+".log")); err != nil {
		fmt.Println("Error initializing logger:", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Info("Logger initialized successfully")
	logger.Print(ui.Logo)

	err = rootCmd.Execute()
	if err != nil {
		logger.Error("Error executing root command: %v", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(
		&cfgFile, "config", "config.yml", "config file (default is $HOME/."+projectName+"/config.yml)",
	)

	rootCmd.Flags().StringVarP(&path, "path", "p", "", "Path to open")
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
