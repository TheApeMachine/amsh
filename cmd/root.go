package cmd

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/buffer"
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
	cfgFile string
	path    string

	rootCmd = &cobra.Command{
		Use:   "amsh",
		Short: "A minimal shell and vim-like text editor with A.I. capabilities",
		Long:  roottxt,
		RunE: func(cmd *cobra.Command, args []string) error {
			prog := tea.NewProgram(
				buffer.New(path),
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
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(
		&cfgFile, "config", "config.yml", "config file (default is $HOME/.data/config.yml)",
	)

	rootCmd.Flags().StringVarP(&path, "path", "p", "", "Path to open")
}

func initConfig() {
	var err error

	if err = writeConfig(); err != nil {
		log.Fatal(err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("$HOME/.data")

	if err = viper.ReadInConfig(); err != nil {
		log.Println("failed to read config file", err)
		return
	}
}

func writeConfig() (err error) {
	var (
		home, _ = os.UserHomeDir()
		fh      fs.File
		buf     bytes.Buffer
	)

	fullPath := home + "/.data/" + cfgFile

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

	if err = os.Mkdir(home+"/.data", os.ModePerm); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err = os.WriteFile(fullPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	log.Println("wrote config file to", fullPath)
	return
}

const roottxt = `amsh v0.0.1
A minimal shell and vim-like text editor written in Go, with integrated A.I. capabilities.
Different from other A.I. integrations, it uses multiple A.I. models that engage independently
in conversation with each other and the user, improving the developer experience and providing
a more human-like interaction.
`
