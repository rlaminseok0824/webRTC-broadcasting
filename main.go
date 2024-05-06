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
		// Client에서는 자신의 LocalDescription을 보내고, 서버에서는 이를 바탕으로 answer를 만들고 이와 관련된 localDescription 보내 연결 설정을 한다.
		case "offer":
			offer := webrtc.SessionDescription{}
			Decode(message.Data, &offer, false)
			if(isBroadcast){
				go Broadcast(offer)
			}else {
				go View(offer)
			}
			localDescription := <- LocalDescriptionChan
			if writeErr := c.WriteJSON(&websocketMessage{
				Event : "lsp",
				Data: localDescription,
			}); writeErr != nil {
				log.Println("writeErr : ", writeErr)
			}
		case "candidate":
		case "answer":
		}
	}
}