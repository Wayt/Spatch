package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type User struct {
	Name           string   `yaml:"user"`
	Groups         []string `yaml:"groups"`
	Password       string   `yaml:"password"`
	AuthorizedKeys []string `yaml:"authorized_keys"`
}

var users = map[string]User{}

func loadUsers(filepath string) error {

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	var usersArray []User
	if err := yaml.Unmarshal(data, &usersArray); err != nil {
		return err
	}

	for _, u := range usersArray {
		users[u.Name] = u
	}

	log.Println("loaded ", len(users), "user(s)")

	return nil
}

func authUserPassword(username, password string) bool {

	u, ok := users[username]
	if !ok {
		return false
	}

	return u.Password == fmt.Sprintf("%x", sha1.Sum([]byte(password)))
}

func authUserPublicKey(username string, key ssh.PublicKey) bool {

	u, ok := users[username]
	if !ok {
		return false
	}

	for _, userKeyData := range u.AuthorizedKeys {

		userKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(userKeyData))
		if err != nil {
			log.Println("authUserPublicKey:", username, err)
			continue
		}

		if userKey.Type() == key.Type() && bytes.Compare(userKey.Marshal(), key.Marshal()) == 0 {
			return true
		}
	}

	return false
}

func (u User) AuthorizedEndpoints() []Endpoint {

	var authorized []Endpoint
	for _, e := range endpoints {
		if e.AuthorizedFor(u) {
			authorized = append(authorized, e)
		}
	}

	return authorized
}
