package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"strconv"
)

const (
	TERM_PROMPT    = ">"
	MSG_BUFFER     = 50
	MAX_MSG_LENGTH = 512
)

type Client struct {
	conn       *ssh.ServerConn
	term       *terminal.Terminal
	termWidth  int
	termHeight int
	msg        chan string
}

func NewClient(sshConn *ssh.ServerConn) *Client {

	return &Client{
		conn: sshConn,
		term: nil,
		msg:  make(chan string, MSG_BUFFER),
	}
}

// Write writes the given message
func (c *Client) Write(msg string) {
	c.term.Write([]byte(msg + "\r\n"))
}

// WriteLines writes multiple messages
func (c *Client) WriteLines(msg []string) {
	for _, line := range msg {
		c.Write(line)
	}
}

// Send sends the given message
func (c *Client) Send(msg string) {
	if len(msg) > MAX_MSG_LENGTH {
		return
	}
	select {
	case c.msg <- msg:
	default:
		log.Printf("Msg buffer full, dropping: %s\n", c.conn.RemoteAddr())
		c.conn.Close()
	}
}

// SendLines sends multiple messages
func (c *Client) SendLines(msg []string) {
	for _, line := range msg {
		c.Send(line)
	}
}

func (c *Client) Resize(width, height int) error {
	width = 1000000 // TODO: Remove this dirty workaround for text overflow once ssh/terminal is fixed
	err := c.term.SetSize(width, height)
	if err != nil {
		log.Printf("Resize failed: %dx%d\n", width, height)
		return err
	}
	c.termWidth, c.termHeight = width, height
	return nil
}

func (c *Client) handleShell(channel ssh.Channel) {

	defer channel.Close()

	go func() {
		// Block until done, then remove.
		c.conn.Wait()
		close(c.msg)
	}()

	go func() {
		for msg := range c.msg {
			c.Write(msg)
		}
	}()

	c.Send("Welcome to Spatch !")
	c.Send("Choose a server:")

	for i, host := range hosts {
		c.Send(fmt.Sprintf("\t(%d) %s", i, host))
	}

	for {
		line, err := c.term.ReadLine()
		if err != nil {
			break
		}
		log.Println("received:", line)

		num, err := strconv.ParseInt(line, 10, 32)
		if err != nil {
			c.Send(err.Error())
			continue
		}

		if num < 0 || int(num) >= len(hosts) {
			c.Send("invalid choice")
			continue
		}

		c.Send("connecting to " + hosts[num])
	}
}

func (c *Client) handleChans(channels <-chan ssh.NewChannel) {

	hasShell := false

	// Service the incoming Channel channel.
	for newChannel := range channels {
		// Channels have a type, depending on the application level
		// protocol intended. In the case of a shell, the type is
		// "session" and ServerShell may be used to present a simple
		// terminal interface.
		if newChannel.ChannelType() != "session" {
			log.Println("invalid channel type:", newChannel.ChannelType())
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Println("could not accept channel:", err)
			continue
		}
		defer channel.Close()

		c.term = terminal.NewTerminal(channel, TERM_PROMPT)

		for req := range requests {
			var width, height int
			var ok bool

			switch req.Type {
			case "shell":
				if c.term != nil && !hasShell {
					go c.handleShell(channel)
					ok = true
					hasShell = true
				}
			case "pty-req":
				width, height, ok = parsePtyRequest(req.Payload)
				if ok {
					err := c.Resize(width, height)
					ok = err == nil
				}
				// case "window-change":
				// 	width, height, ok = parseWinchRequest(req.Payload)
				// 	if ok {
				// 		err := c.Resize(width, height)
				// 		ok = err == nil
				// 	}
			}

			if req.WantReply {
				req.Reply(ok, nil)
			}
		}
	}
}
