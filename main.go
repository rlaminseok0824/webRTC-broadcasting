package main

import (
	"encoding/json"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/pion/webrtc/v4"
)

var (
	// webRTC의 기본 설정으로 default 구글 strun 서버로 ICE 서버를 설정
	defaultPeerConnectionConfig = webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	trackLocals map[string]*webrtc.TrackLocalStaticRTP
	localTrackChan = make(chan *webrtc.TrackLocalStaticRTP)
	LocalDescriptionChan = make(chan string)
)

func main(){
	log.SetFlags(0)
	app := fiber.New()

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c){
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/:room", websocket.New(wsHandler))


	log.Fatal(app.Listen(":3000"))
}

func wsHandler(c *websocket.Conn) {
	// ex) ws://localhost:3000/ws/123?isBroadcast=true
	log.Println(c.Params("room")) //123
	isBroadcast := (c.Query("isBroadcast") == "true") //true

	message := &websocketMessage{}
	for {
		_, msg, readErr := c.ReadMessage()
		if readErr != nil {
			log.Println("readErr : ", readErr)
		} else if err := json.Unmarshal(msg, &message); err != nil {
			log.Println("json.Unmarshal : ", err)
		}

		switch message.Event {
		case "offer":
			offer := webrtc.SessionDescription{}
			Decode(message.Data, &offer, false)
			if(isBroadcast){
				go Broadcast(offer)
			}
		case "candidate":
		case "answer":
		}
	}
}