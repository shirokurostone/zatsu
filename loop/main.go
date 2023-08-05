package main

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"os/signal"
	"time"
)

var rootCmd = &cobra.Command{
	Use:  "[flags] command",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		sleep, err := cmd.Flags().GetString("sleep")
		if err != nil {
			return err
		}
		timestamp, err := cmd.Flags().GetBool("timestamp")
		if err != nil {
			return err
		}
		noColor, err := cmd.Flags().GetBool("no-color")
		if err != nil {
			return err
		}
		noPrompt, err := cmd.Flags().GetBool("no-prompt")
		if err != nil {
			return err
		}
		counter, err := cmd.Flags().GetString("counter")
		if err != nil {
			return err
		}
		exitOnFailure, err := cmd.Flags().GetBool("exit-on-failure")
		if err != nil {
			return err
		}
		cmdline := args[0]

		var duration time.Duration
		if duration, err = time.ParseDuration(sleep); err != nil {
			return err
		}
		if noColor {
			color.NoColor = true
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

	CommandLoop:
		for i := 0; ; i++ {

			select {
			case <-time.After(duration):
				break
			case <-ctx.Done():
				break CommandLoop
			}

			time.Sleep(duration)

			if !noPrompt {
				if timestamp {
					color.Green("[%s]$ %s\n", time.Now().Format(time.RFC3339), cmdline)
				} else {
					color.Green("$ %s\n", cmdline)
				}
			}
			command := exec.Command("bash", "-c", cmdline)
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
			command.Env = append(command.Environ(), fmt.Sprintf("%s=%d", counter, i))
			if err = command.Start(); err != nil {
				if exitOnFailure {
					return nil
				}
			}

			ch := make(chan error, 1)
			go func() {
				defer close(ch)
				ch <- command.Wait()
			}()

			select {
			case e := <-ch:
				if e != nil {
					if exitOnFailure {
						return nil
					}
					continue CommandLoop
				}
				break
			case <-ctx.Done():
				if err := command.Process.Signal(os.Interrupt); err != nil {
					return nil
				}
				break CommandLoop
			}
		}
		return nil
	},
}

func init() {
	rootCmd.Flags().StringP("sleep", "s", "1s", "sleep time")
	rootCmd.Flags().BoolP("timestamp", "t", false, "show timestamp")
	rootCmd.Flags().Bool("no-color", false, "disable color output")
	rootCmd.Flags().Bool("no-prompt", false, "disable prompt")
	rootCmd.Flags().String("counter", "i", "counter variable name")
	rootCmd.Flags().BoolP("exit-on-failure", "e", false, "exit if command failed")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
