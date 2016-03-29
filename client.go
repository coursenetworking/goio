package goio

import (
	"fmt"
	"sync"
	"time"

	"gopkg.in/mgo.v2/bson"
)

func NewClientId() string {
	return bson.NewObjectId().Hex()
}

func NewClient(user *User) *Client {
	clt := &Client{
		Id:            NewClientId(),
		User:          user,
		message:       make(chan *Message),
		Messages:      make([]*Message, 0, 10),
		close:         make(chan struct{}),
		fetchMessages: make(chan struct{}),
		msgs:          make(chan []*Message),
		lastHandshake: time.Now().Unix(),
	}

	Clients().addClt <- clt
	user.addClt <- clt

	// check for client life cycle
	go func(clt *Client) {
		for {
			time.Sleep(time.Duration(LifeCycle) * time.Second) // never timeout if any message come

			if time.Now().Unix()-clt.lastHandshake < LifeCycle {
				continue
			}

			// timeout, kill this client
			fmt.Printf("user %s clt %s timeout\n", clt.User.Id, clt.Id)
			clt.User.DelClt <- clt

			clt.Messages = nil
			clt.User = nil

			close(clt.close)
			fmt.Printf("---> clt %s deleted, break life cycle loop\n", clt.Id)
			return
		}
	}(clt)

	// listen for client message
	go func(clt *Client) {
		for {
			select {
			case msg := <-clt.message:
				clt.Messages = append(clt.Messages, msg)

			case <-clt.fetchMessages:
				clt.msgs <- clt.Messages
				clt.Messages = make([]*Message, 0, 10)

			case <-clt.close:
				fmt.Printf("---> clt %s del, break listen loop\n", clt.Id)
				// break
				return
			}
		}
	}(clt)

	fmt.Printf("new client %s to user %s\n", clt.Id, clt.User.Id)
	return clt
}

type Client struct {
	Id            string
	User          *User
	message       chan *Message
	Messages      []*Message
	fetchMessages chan struct{}
	close         chan struct{}
	msgs          chan []*Message
	lastHandshake int64
	lock          sync.RWMutex
}

func (c *Client) Handshake() {
	c.lastHandshake = time.Now().Unix()
}

func (c *Client) Msgs() []*Message {
	c.fetchMessages <- struct{}{}
	return <-c.msgs
}
