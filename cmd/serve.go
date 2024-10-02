package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/service"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the amsh service",
	Long:  serveTxt,
	RunE: func(cmd *cobra.Command, _ []string) error {
		srv := service.NewHTTPS()

		// Graceful shutdown setup
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		go func() {
			if errnie.Error(srv.Up()) != nil {
				return
			}
		}()

		<-ctx.Done()
		stop()
		fmt.Println("Shutting down server...")

		if err := srv.Shutdown(); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}

		fmt.Println("Server gracefully stopped")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

const serveTxt = `
Starts a fiber 3 HTTP service, using pre-forking to handle requests,
benefiting from additional performance and reduced memory usage due to
Go not having to manage memory across multiple processes.
`
