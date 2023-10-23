package core

import (
	"os"
	"os/signal"
	"syscall"
	"yajs/config"
	"yajs/utils"
)

func trapSignals() {
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)

		for sig := range sigchan {
			switch sig {
			case syscall.SIGQUIT:
				utils.Logger.Warningf("[INFO] SIGQUIT: Quitting process immediately")
				os.Exit(0)

			case syscall.SIGTERM:
				utils.Logger.Warningf("[INFO] SIGTERM: Shutting down servers then terminating")
				os.Exit(0)

			case syscall.SIGUSR1 , syscall.SIGUSR2:
				utils.Logger.Warningf("Receive %s : Reloading",sig)
				config.Reload()
				utils.Logger.Warningf("Receive %s : Reload finish",sig)
			case syscall.SIGHUP:
			}
		}
	}()
}

