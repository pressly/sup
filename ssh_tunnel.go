package sup

import (
	"github.com/pkg/errors"
	"fmt"
	"net"
	"io"
	"strings"
	"os"
)

type SSHTunnel struct {
	Tunnel Tunnel
	SSHClient *SSHClient
	err chan error
}

func (t Tunnel) String() string {
	return fmt.Sprintf("%d:%s:%d\n", t.ListenPort, t.Host,  t.DstPort)
}

func (t *SSHTunnel) StartTunnel() {
	lAddr := fmt.Sprintf("localhost:%d", t.Tunnel.ListenPort)
	ln, err := t.SSHClient.conn.Listen("tcp", lAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Remote tunnel listen failed: %s\n", err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Remote tunnel accept faild: %s\n", err)
		}
		go t.forward(conn)
	}
}

func (t *SSHTunnel) forward(remoteConn net.Conn) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d",  t.Tunnel.Host, t.Tunnel.DstPort))
	if err != nil {
		t.err <- errors.Wrap(err, "local tunnel connection error")
	}

	copy := func(writer, reader net.Conn) {
		_, err:= io.Copy(writer, reader)
		if err != nil && ! strings.Contains(err.Error(), "use of closed network connection")  {
			t.err <-  errors.Wrap(err, "Tunnel forward error")
		}
		writer.Close()
	}

	go copy(conn, remoteConn)
	go copy(remoteConn, conn);
}
