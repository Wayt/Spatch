package main

import (
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"log"
)

type Session struct {
	*ssh.Session
	closed bool
	c      *Client
	Addr   string
}

func NewSession(c *Client, addr string) (*Session, error) {

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	sshSession, err := c.client.NewSession()
	if err != nil {
		log.Println("client.NewSession:", err)
		return nil, err
	}

	session := &Session{
		sshSession,
		false,
		c,
		addr,
	}

	if err := session.pipeToTerm(c.term); err != nil {
		c.Send(err.Error())
		return nil, err
	}

	log.Println("Open remote shell...")

	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	// Request pseudo terminal
	if err := session.RequestPty("xterm", c.termHeight, c.termWidth, modes); err != nil {
		log.Println("request for pseudo terminal failed:", err)
		return nil, err
	}

	if err = session.Shell(); err != nil {
		log.Println("session.Run:", err)
		return nil, err
	}

	go func() {
		session.Wait()
		session.closed = true
	}()

	return session, nil
}

func (s *Session) pipeToTerm(term *terminal.Terminal) error {

	if err := s.InPipe(term); err != nil {
		return err
	}

	if err := s.OutPipe(term); err != nil {
		return err
	}

	if err := s.ErrPipe(term); err != nil {
		return err
	}

	return nil
}

func (s *Session) OutPipe(term *terminal.Terminal) error {

	out, err := s.StdoutPipe()
	if err != nil {
		return err
	}

	go func() {

		if _, err := io.Copy(term, out); err != nil {

			log.Println("io.Copy(term, out:", err)
		}
	}()

	return nil
}

func (s *Session) ErrPipe(term *terminal.Terminal) error {

	out, err := s.StderrPipe()
	if err != nil {
		return err
	}

	go func() {

		if _, err := io.Copy(term, out); err != nil {

			log.Println("io.Copy(term, out:", err)
		}
	}()

	return nil
}

func (s *Session) InPipe(term *terminal.Terminal) error {

	in, err := s.StdinPipe()
	if err != nil {
		return err
	}

	// Copy input command to remote ssh
	go func() {
		defer in.Close()
		for !s.closed {

			line, err := term.ReadLine()
			if err != nil {

				if err == io.EOF {
					s.Close()
					return
				}
				log.Println("term.ReadLine:", err)
				return
			}

			CMDLogln(s.c.User.Name, s.Addr, line)

			if _, err := in.Write([]byte(line + "\n")); err != nil {
				log.Println("term.ReadLine:", err)
				return
			}
		}
	}()
	return nil
}
