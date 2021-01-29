// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package logs_routes

import (
	"errors"
	"net"
	"net/http"
	"strings"

	"gitlab.com/dataptive/styx/api"

	"github.com/gorilla/websocket"
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

func UpgradeWebsocket(w http.ResponseWriter, r *http.Request, allowedOrigins []string)  (conn *websocket.Conn, err error) {

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) (ret bool) {
			origin := r.Header.Get("Origin")
			return matchOrigin(origin, allowedOrigins)
		},
	}

	conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func matchOrigin(origin string, allowed []string) (match bool) {

	if origin == "" {
		return false
	}

	for _, a := range allowed {
		if matchWildcard(origin, a) {
			return true
		}
	}

	return false
}

func matchWildcard(s string, pattern string) (match bool) {

	index := strings.IndexByte(pattern, '*')

	if index == -1 {
		if s == pattern {
			return true
		}

		return false
	}

	prefix := pattern[0:index]
	suffix := pattern[index+1:]

	if !strings.HasPrefix(s, prefix) {
		return false
	}
	if !strings.HasSuffix(s, suffix) {
		return false
	}

	return true
}
