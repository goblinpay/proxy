// from https://dave.cheney.net/2010/10/05/how-to-dial-remote-ssltls-services-in-go

package main

import (
	"crypto/tls"
	"net"
)

var config = &tls.Config{
	Rand:    nil, // uses random reader in package crypto/rand
	Time:    nil, // uses time.Now
	RootCAs: nil, // uses the host's root CA set
}

func dialTLS(raddr *net.TCPAddr) (c *tls.Conn, err error) {

	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		return nil, err
	}

	c = tls.Client(conn, config)
	err = c.Handshake()
	if err != nil {
		c.Close()
		return nil, err
	}

	return c, nil
}
