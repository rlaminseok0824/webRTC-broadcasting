package handler

type websocketMessage struct {
	Event string `json:"event"`
	Data  string `json:"data"`
	ID   string `json:"id"`
}