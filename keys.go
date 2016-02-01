package main

import (
	"bufio"
	"bytes"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
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

var authorizedKeys []ssh.PublicKey

func loadAuthorizedKeys(filepath string) error {

	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		log.Println(scanner.Text())
		key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(scanner.Text()))
		if err != nil {
			log.Fatal(err)
		}

		authorizedKeys = append(authorizedKeys, key)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	log.Println("loaded ", len(authorizedKeys), "authorized key(s)")

	return nil
}

func validatePublicKey(key ssh.PublicKey) bool {

	for _, k := range authorizedKeys {
		if k.Type() == key.Type() && bytes.Compare(k.Marshal(), key.Marshal()) == 0 {
			return true
		}
	}
	return false
}
