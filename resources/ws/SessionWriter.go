package ws

import melody "gopkg.in/olahol/melody.v1"

// SessionWriter helps to serialize a spesific session json and respond all data
type SessionWriter struct {
	Session *melody.Session
}

func (w SessionWriter) Write(p []byte) (int, error) {
	w.Session.Write(p)
	return len(p), nil
}
