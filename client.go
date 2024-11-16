package main

import (
	"database/sql"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types/events"
)

type WhatsappClient struct {
	waClient *whatsmeow.Client
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

	case *events.LoggedOut:
		err := c.waClient.Logout()
		if err != nil {
			log.Err(err)
		}
		c.db.Exec("DELETE FROM users WHERE token = ?", c.userId)
		c.logoutCh <- true

		state.Lock()
		delete(state.clients, c.userId)
		state.Unlock()
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
		parsedJid, _ := parseJID(*jid)
		deviceStore, _ := container.GetDevice(parsedJid)
		waClient = whatsmeow.NewClient(deviceStore, nil)
	}

	logOutCh := make(chan bool)
	client := &WhatsappClient{
		waClient,
		logOutCh,
		db,
		userId,
		userName,
	}
	client.waClient.AddEventHandler(client.EventHandler)
	return client
}
