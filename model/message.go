package model

import (
	"database/sql"
	"errors"
)

type Message struct {
	ID          string
	SenderID    string
	ReceiverID  string
	DateTime    string
	TextContent string
	HasReceived bool
}

func ReadMessage(rows *sql.Rows) (Message, error) {
	var item Message
	if err := rows.Scan(
		&item.ID,
		&item.SenderID,
		&item.ReceiverID,
		&item.DateTime,
		&item.TextContent,
		&item.HasReceived,
	); err != nil {
		return Message{}, errors.New("Message Convert Error " + err.Error())
	}
	return item, nil
}

type MessageContact struct {
	SenderID    string
	Username    string
	TextContent string
	DateTime    string
	HasUnread   bool
}

func ReadMessageContact(rows *sql.Rows) (MessageContact, error) {
	var item MessageContact
	if err := rows.Scan(
		&item.SenderID,
		&item.Username,
		&item.TextContent,
		&item.DateTime,
		&item.HasUnread,
	); err != nil {
		return MessageContact{}, errors.New("MessageContact Convert Error " + err.Error())
	}
	return item, nil
}
