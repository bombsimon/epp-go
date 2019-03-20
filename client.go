package epp

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/bombsimon/epp-go/types"
)

type Client struct {
	TLSConfig *tls.Config
	messages  chan []byte
	conn      net.Conn
}

func (c *Client) Connect(server string) ([]byte, error) {
	if c.TLSConfig == nil {
		c.TLSConfig = &tls.Config{}
	}
	conn, err := tls.Dial("tcp", server, c.TLSConfig)
	if err != nil {
		return nil, err
	}

	// Read the greeting.
	greeting, err := ReadMessage(conn)
	if err != nil {
		return nil, err
	}

	c.conn = conn

	return greeting, nil
}

func (c *Client) Send(data []byte) ([]byte, error) {
	err := WriteMessage(c.conn, data)
	if err != nil {
		_ = c.conn.Close()
		return nil, err
	}

	_ = c.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	msg, err := ReadMessage(c.conn)
	if err != nil {
		_ = c.conn.Close()
		return nil, err
	}

	return msg, nil
}

// Login will perform a login to an EPP server.
func (c *Client) Login(username, password string) ([]byte, error) {
	login := types.Login{
		ClientID: username,
		Password: password,
		Options: types.LoginOptions{
			Version:  "1.0",
			Language: "en",
		},
		Services: types.LoginServices{
			ObjectURI: []string{
				"urn:ietf:params:xml:ns:domain-1.0",
				"urn:ietf:params:xml:ns:contact-1.0",
				"urn:ietf:params:xml:ns:host-1.0",
			},
			ServiceExtension: types.LoginServiceExtension{
				ExtensionURI: []string{
					"urn:ietf:params:xml:ns:secDNS-1.0",
					"urn:ietf:params:xml:ns:secDNS-1.1",
				},
			},
		},
	}

	encoded, err := Encode(login, ClientXMLAttributes(), "")
	if err != nil {
		return nil, err
	}

	return c.Send(encoded)
}
