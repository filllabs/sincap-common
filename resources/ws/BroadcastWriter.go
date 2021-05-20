package ws

import melody "gopkg.in/olahol/melody.v1"

// BroadcastWriter helps to serialize metrics json and broadcast all data
type BroadcastWriter struct {
	Socket *melody.Melody
}

func (w BroadcastWriter) Write(p []byte) (int, error) {
	w.Socket.Broadcast(p)
	return len(p), nil
}
