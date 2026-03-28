package cli

import (
	"io"
	"net/http"

	"github.com/adrg/xdg"
	"github.com/maximerivest/mcp2cli/internal/daemon"
	"github.com/maximerivest/mcp2cli/internal/elicitation"
	mcpclient "github.com/maximerivest/mcp2cli/internal/mcp/client"
)

func terminalConnectOptions(in io.Reader, errOut io.Writer) mcpclient.ConnectOptions {
	handler := elicitation.NewHandler(in, errOut)
	return mcpclient.ConnectOptions{
		ElicitationHandler: handler.Handle,
		DaemonCheck:        daemonCheck,
	}
}

func daemonCheck(serverName string) (*http.Client, string, bool) {
	if !daemon.IsRunning(xdg.DataHome, serverName) {
		return nil, "", false
	}
	return daemon.DialSocket(xdg.DataHome, serverName), "http://unix/" + serverName, true
}
