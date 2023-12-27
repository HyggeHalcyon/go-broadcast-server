package server

import (
	"io"
	"log"
	"net"
)

const MSG_SIZE = 1024 * 2

type Client struct {
	Conn net.Conn
	pipe chan Payload
	quit chan Client
}

func NewClient(conn net.Conn, pipe chan Payload, quit chan Client) Client {
	return Client{
		Conn: conn,
		pipe: pipe,
		quit: quit,
	}
}

func (c *Client) Start() {
	go func() {
		// close connection and send the client to the channel
		// so it will be removed from the pool
		defer func() {
			c.quit <- *c
			c.Conn.Close()
		}()

		buffer := make([]byte, MSG_SIZE)

		for {
			n, err := c.Conn.Read(buffer)
			if err != nil {
				if err.Error() == io.EOF.Error() { // ignore, only happened because client disconnected
					return
				} else {
					log.Printf("failed reading from %s: %s\n", c.toString(), err.Error())
					return
				}
			}

			// send message to the broadcasting thread
			c.pipe <- Payload{
				Source:  *c,
				Message: buffer[:n],
			}
		}
	}()
}

func (c *Client) toString() string {
	return c.Conn.RemoteAddr().String()
}
