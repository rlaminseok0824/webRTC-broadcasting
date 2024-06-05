package handler

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/pion/webrtc/v4"

	"github.com/webRTC-broadcasting/utils"
)

var (
	ListLock sync.RWMutex
	// webRTC의 기본 설정으로 default 구글 stun 서버로 ICE 서버를 설정
	defaultPeerConnectionConfig = webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	TrackLocals map[string]*webrtc.TrackLocalStaticRTP
	localTrackChan = make(chan *webrtc.TrackLocalStaticRTP)
	LocalDescriptionChan = make(chan string)
)

func WsUpgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c){
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

func WsHandler(c *websocket.Conn) {
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
			// Client에서는 자신의 LocalDescription을 보내고, 서버에서는 이를 바탕으로 answer를 만들고 이와 관련된 localDescription 보내 연결 설정을 한다.
			offer := webrtc.SessionDescription{}
			utils.Decode(message.Data, &offer, false)

			if(isBroadcast){
				go Broadcast(offer,message.ID)
			}else {
				go View(offer,message.ID)
			}
			localDescription := <- LocalDescriptionChan
			log.Println("Create Local Description Based on Offer : \n", localDescription)
			if writeErr := c.WriteJSON(&websocketMessage{
				Event : "lsp",
				Data: localDescription,
			}); writeErr != nil {
				log.Println("writeErr : ", writeErr)
			}
		case "candidate":
		case "answer":
		case "track":
			// Client에서는 자신의 서버가 시작되었음을 서버에 알린다.
			ListLock.Lock()
			localTrack := <- localTrackChan
			TrackLocals[localTrack.ID()] = localTrack //임의로 video로 설정 추후 바꿈
			ListLock.Unlock()
			log.Println("Track added ", localTrack)

			if writeErr := c.WriteJSON(&websocketMessage{
				Event : "track",
				Data: localTrack.ID(),
			}); writeErr != nil {
				log.Println("writeErr : ", writeErr)
			}
		}
	}
}