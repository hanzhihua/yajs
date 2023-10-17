package utils

import (
	"github.com/op/go-logging"
	"os"
)

var Logger *logging.Logger

func init() {
	Logger = logging.MustGetLogger("NAME")
	format := logging.MustStringFormatter(
		`%{color}%{time:2006-01-02T15:04:05.000} %{shortfile} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	stdoutHandler := logging.NewBackendFormatter(logging.NewLogBackend(os.Stdout, "", 0), format)
	logging.SetBackend(stdoutHandler)
}
