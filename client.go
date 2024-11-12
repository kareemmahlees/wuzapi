package main

import (
	"database/sql"
	"fmt"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types/events"
)

type WhatsappClient struct {
	waClient *whatsmeow.Client
	logInCh  chan bool
	logoutCh chan bool
	db       *sql.DB
	userId   string
	userName string
}

func (c *WhatsappClient) EventHandler(rawEvent any) {
	switch event := rawEvent.(type) {
	case *events.PairSuccess:
		jid := event.ID
		c.db.Exec("INSERT INTO users (name, token, webhook, expiration, events, jid, qrcode) VALUES (?, ?, ?, ?, ?, ?, ?)",
			c.userName, c.userId, "", 0, "Message", jid, "")
		err := c.waClient.Connect()
		if err != nil {
			log.Err(err)
		}

		state.Lock()
		state.clients[c.userId] = c
		state.Unlock()

		c.logInCh <- true
	case *events.LoggedOut:
		c.db.Exec("DELETE FROM users WHERE token = ?", c.userId)
		err := c.waClient.Logout()
		if err != nil {
			log.Err(err)
		}

		state.Lock()
		delete(state.clients, c.userId)
		state.Unlock()

		c.logoutCh <- true
	}

}

func NewWhatsappClient(userId string, userName string, jid *string, db *sql.DB) *WhatsappClient {
	appName := "Thyab"
	store.DeviceProps.Os = &appName

	var waClient *whatsmeow.Client

	if jid == nil {
		deviceStore := container.NewDevice()
		waClient = whatsmeow.NewClient(deviceStore, nil)
	} else {
		fmt.Println(*jid)
		parsedJid, _ := parseJID(*jid)
		deviceStore, _ := container.GetDevice(parsedJid)
		waClient = whatsmeow.NewClient(deviceStore, nil)
	}

	loggedInCh := make(chan bool, 1)
	logInCh := make(chan bool, 1)
	client := &WhatsappClient{
		waClient,
		loggedInCh,
		logInCh,
		db,
		userId,
		userName,
	}
	client.waClient.AddEventHandler(client.EventHandler)
	return client
}
