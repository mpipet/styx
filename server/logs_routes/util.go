package logs_routes

import (
	"errors"
	"net"
	"net/http"

	"gitlab.com/dataptive/styx/api"
)

var (
	ErrUnsupportedUpgrade    = errors.New("server: protocol doesn't support connection upgrade")
	ErrDataSentBeforeUpgrade = errors.New("server: client sent data before upgrade completion")
)

func UpgradeTCP(w http.ResponseWriter) (c *net.TCPConn, err error) {

	hj, ok := w.(http.Hijacker)
	if !ok {
		err := ErrUnsupportedUpgrade
		api.WriteError(w, http.StatusInternalServerError, err)
		return nil, err
	}
	header := w.Header()
	header.Add("Connection", "Upgrade")
	header.Add("Upgrade", api.StyxProtocolString)

	api.WriteResponse(w, http.StatusSwitchingProtocols, nil)

	conn, bufrw, err := hj.Hijack()
	if err != nil {
		conn.Close()
		return nil, err
	}

	if bufrw.Reader.Buffered() > 0 {
		conn.Close()
		err := ErrDataSentBeforeUpgrade
		return nil, err
	}

	return conn.(*net.TCPConn), nil
}
