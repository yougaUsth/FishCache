package service

import "github.com/satori/go.uuid"

type Session struct {
	sID      string
	uID      string
	conn     *Conn
	settings map[string]interface{}
}

func NewSession(conn *Conn) *Session {
	id, _ := uuid.NewV4()
	session := &Session{
		sID:      id.String(),
		uID:      "",
		conn:     conn,
		settings: make(map[string]interface{}),
	}
	return session
}
