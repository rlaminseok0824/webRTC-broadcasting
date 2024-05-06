package main

import (
	"encoding/json"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
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
	// ex) ws://localhost:3000/ws/123
	log.Println(c.Params("room")) //123

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
		case "candidate":
		case "answer":
		}
	}
}