package main

import (
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
)

var bind = flag.String("bind", ":8080", "SSH bind address")
var rsaKey = flag.String("rsa_key", "keys/ssh_host_rsa_key", "RSA key path")
var dsaKey = flag.String("dsa_key", "keys/ssh_host_dsa_key", "DSA key path")
var privateKey = flag.String("ssh_key", "keys/id_rsa", "RSA private key path")
var endpointsFile = flag.String("endpts", "endpoints.yml", "Endpoints configuration file")

var config = &ssh.ServerConfig{
	NoClientAuth: false,
	PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
		// Should use constant-time compare (or better, salt+hash) in
		// a production setting.
		if c.User() == "test" && string(pass) == "test" {
			return nil, nil
		}
		return nil, fmt.Errorf("password rejected for %q", c.User())
	},
}

func readHostKeys() {

	// Read RSA key
	{
		privateBytes, err := ioutil.ReadFile(*rsaKey)
		if err != nil {
			log.Fatal(err)
		}

		private, err := ssh.ParsePrivateKey(privateBytes)
		if err != nil {
			log.Fatal(err)
		}

		config.AddHostKey(private)
	}

	// Read DSA key
	{
		privateBytes, err := ioutil.ReadFile(*dsaKey)
		if err != nil {
			log.Fatal(err)
		}

		private, err := ssh.ParsePrivateKey(privateBytes)
		if err != nil {
			log.Fatal(err)
		}

		config.AddHostKey(private)
	}
}

func main() {

	flag.Parse()

	log.Println("Read host keys...")
	readHostKeys()

	log.Println("Load endpoints...")
	loadEndpoints(*endpointsFile)

	log.Println("running spatch on", *bind)

	ln, err := net.Listen("tcp", *bind)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			return
		}

		go handle(conn)
	}
}

func handle(conn net.Conn) {

	defer conn.Close()

	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Connection from %s, %s, %s\n", sshConn.RemoteAddr(), sshConn.User(), sshConn.ClientVersion())

	go ssh.DiscardRequests(reqs)

	client := NewClient(sshConn)
	client.handleChans(chans)
}
