package main

import (
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var bind = flag.String("bind", ":8080", "SSH bind address")
var rsaKey = flag.String("rsa_key", "keys/ssh_host_rsa_key", "RSA key path")
var dsaKey = flag.String("dsa_key", "keys/ssh_host_dsa_key", "DSA key path")
var privateKey = flag.String("ssh_key", "keys/id_rsa", "RSA private key path")
var endpointsFile = flag.String("endpts", "endpoints.yml", "Endpoints configuration file")
var usersFile = flag.String("users", "users.yml", "Users configuration file")
var commandLogFile = flag.String("cmd_log", "stdout", "Commands log file")

var config = &ssh.ServerConfig{
	NoClientAuth: false,
	PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {

		if !authUserPassword(c.User(), string(pass)) {
			return nil, fmt.Errorf("password rejected for %q", c.User())
		}

		return nil, nil
	},
	PublicKeyCallback: func(c ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {

		if !authUserPublicKey(c.User(), key) {
			return nil, fmt.Errorf("ssh key rejected for %q", c.User())
		}
		return nil, nil
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

var sighup = make(chan os.Signal, 1)

func loadConfig() {

	log.Println("Load endpoints...")
	loadEndpoints(*endpointsFile)

	log.Println("Load users...")
	loadUsers(*usersFile)
}

func main() {

	flag.Parse()

	log.Println("Read host keys...")
	readHostKeys()

	openLogFile(*commandLogFile)

	loadConfig()

	// register signal handler
	signal.Notify(sighup, syscall.SIGHUP)

	go func() {
		for range sighup {

			log.Println("Re-loading configuration")
			loadConfig()
		}
	}()

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

	defer func() {

		if err := recover(); err != nil {
			log.Println("handle: recover:", err)
		}
	}()

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
