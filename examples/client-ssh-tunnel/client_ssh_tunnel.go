// Copyright 2013 Alexandre Fiori
// Copyright 2019 Matthias Endler
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Event Socket client that connects via SSH to FreeSWITCH to originate a new call.
//
// In order for this example to work, you need to create a passphraseless
// ssh key pair.
//
// $ ssh-keygen -b 2048 -t rsa -f /tmp/id_rsa -q -N ""
//
// Then you need to copy the public key from /tmp/id_rsa.pub to the user's
// ~/.ssh/authorized_keys file on the remote host. If you don't copy the
// public key, the password authentication is used instead.
//
// WARNING: Please be aware, that anyone which has the private key, can connect
// to the remote system, without any password!
package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/m12r/go-eventsocket/eventsocket"
	"golang.org/x/crypto/ssh"
)

const dest = "sofia/internal/1000%127.0.0.1"
const dialplan = "&socket(localhost:9090 async)"
const sshAddr = "remoteAddr:22"
const sshUser = "user"
const sshPasswd = "super-secret-passwd"
const sshPrivateKeyPath = "/tmp/id_rsa"
const remoteAddr = "127.0.0.1:8021"

func parsePrivateKey(keyPath string) (ssh.Signer, error) {
	buff, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(buff)
}

func makeSSHConfig(user, password string) (*ssh.ClientConfig, error) {
	key, err := parsePrivateKey(sshPrivateKeyPath)
	if err != nil {
		return nil, err
	}

	config := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return &config, nil
}

func main() {
	sshCfg, err := makeSSHConfig(sshUser, sshPasswd)
	if err != nil {
		log.Fatalln(err)
	}

	sshConn, err := ssh.Dial("tcp", sshAddr, sshCfg)
	if err != nil {
		log.Fatalln(err)
	}

	remote, err := sshConn.Dial("tcp", remoteAddr)
	if err != nil {
		log.Fatalln(err)
	}

	c, err := eventsocket.NewConnection(remote, "ClueCon")
	if err != nil {
		log.Fatalln(err)
	}
	c.Send("events json ALL")
	c.Send(fmt.Sprintf("bgapi originate %s %s", dest, dialplan))
	for {
		ev, err := c.ReadEvent()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("\nNew event")
		ev.PrettyPrint()
		if ev.Get("Answer-State") == "hangup" {
			break
		}
	}
	c.Close()
}
