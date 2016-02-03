package main

import (
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
)

func readSshPrivateKey() ssh.AuthMethod {

	privateBytes, err := ioutil.ReadFile(*privateKey)
	if err != nil {
		log.Fatal(err)
	}

	signer, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal(err)
	}

	return ssh.PublicKeys(signer)
}
