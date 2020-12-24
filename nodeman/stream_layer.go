package nodeman

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"gitlab.com/dataptive/styx/api"

	"github.com/hashicorp/raft"
)

type streamLayer struct {
	address    net.Addr
	acceptChan chan net.Conn
	closeChan  chan struct{}
	closed     bool
}

func newStreamLayer(address net.Addr) (sl *streamLayer) {

	sl = &streamLayer{
		address:    address,
		acceptChan: make(chan net.Conn),
		closeChan:  make(chan struct{}),
		closed:     false,
	}

	return sl
}

func (sl *streamLayer) Accept() (conn net.Conn, err error) {

	if sl.closed {
		return nil, io.EOF
	}

	select {
	case conn = <-sl.acceptChan:
		return conn, nil
	case <-sl.closeChan:
		return nil, io.EOF
	}
}

func (sl *streamLayer) acceptHandler(conn net.Conn) {

	sl.acceptChan <- conn
}

func (sl *streamLayer) Close() (err error) {

	if sl.closed {
		return nil
	}

	sl.closeChan <- struct{}{}

	return nil
}

func (sl *streamLayer) Addr() (addr net.Addr) {

	return sl.address
}

func (sl *streamLayer) Dial(address raft.ServerAddress, timeout time.Duration) (conn net.Conn, err error) {

	conn, err = net.DialTimeout("tcp", string(address), timeout)
	if err != nil {
		return nil, err
	}

	url, err := url.Parse("/nodes")
	if err != nil {
		return nil, err
	}

	req := &http.Request{
		Method:     "GET",
		URL:        url,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}

	req.Header.Set("Upgrade", api.RaftProtocolString)

	err = req.Write(conn)
	if err != nil {
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusSwitchingProtocols {
		return nil, errors.New("nodeman: connection upgrading failed")
	}

	return conn, nil
}
