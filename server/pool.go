package server

import (
	"log"
	"sync"
)

type (
	ThreadSafeClients interface {
		Dispatch()
		Disconnect()
		Push(Client)
		Remove(Client)
		Broadcast(Payload)
	}

	threadSafeClients struct {
		sync.Mutex // Mutual Exclusion Lock, to avoid race condition when accessing data
		clients    []Client
		pipe       chan Payload
		quit       chan Client
	}
)

func NewThreadSafeClients(pipe chan Payload, quit chan Client) ThreadSafeClients {
	return &threadSafeClients{
		pipe: pipe,
		quit: quit,
	}
}

func (cs *threadSafeClients) Dispatch() {
	go func() {
		for {
			payload := <-cs.pipe // blocks, only continues if there's only message to consume
			cs.Broadcast(payload)
		}
	}()
}

func (cs *threadSafeClients) Disconnect() {
	go func() {
		for {
			client := <-cs.quit // blocks, detects when a client disconnects
			cs.Remove(client)
		}
	}()
}

func (cs *threadSafeClients) Push(c Client) {
	cs.Lock()
	defer cs.Unlock()

	log.Printf("new connection from %s\n", c.toString())

	cs.clients = append(cs.clients, c)
}

func (cs *threadSafeClients) Remove(client Client) {
	cs.Lock()
	defer cs.Unlock()

	log.Printf("disconnecting from %s\n", client.toString())

	for i, c := range cs.clients {
		// removing only specific client from the pool
		if c == client {
			cs.clients = append(cs.clients[:i], cs.clients[i+1:]...)
			return
		}
	}
}

func (cs *threadSafeClients) Broadcast(payload Payload) {
	cs.Lock()
	defer cs.Unlock()

	log.Printf("Broadcasting: %s", payload.Message)

	for _, c := range cs.clients {
		// ignore source client
		if payload.Source == c {
			continue
		}

		_, err := c.Conn.Write(payload.Message)
		if err != nil {
			log.Printf("failed broadcasting: %s\n", err.Error())
			continue
		}
	}
}
