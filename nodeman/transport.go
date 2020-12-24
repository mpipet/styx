package nodeman

import (
	"net"
	"time"
)

type transport struct {
	listener net.Listener
}

func newTransport() (t *transport) {

	t = &transport{
		listener: nil,
	}

	return t
}

func (t *transport) Open(addr string) (err error) {

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	t.listener = listener

	return nil
}

func (t *transport) Dial(addr string, timeout time.Duration) (conn net.Conn, err error) {

	dialer := &net.Dialer{
		Timeout: timeout,
	}

	conn, err = dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return conn, err
}

func (t *transport) Accept() (conn net.Conn, err error) {

	conn, err = t.listener.Accept()
	if err != nil {
		return nil, err
	}

	return conn, err
}

func (t *transport) Close() (err error) {

	if t.listener == nil {
		return nil
	}

	err = t.listener.Close()
	if err != nil {
		return err
	}

	return nil
}

func (t *transport) Addr() (addr net.Addr) {

	return t.listener.Addr()
}
