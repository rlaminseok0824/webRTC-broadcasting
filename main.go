package main

import (
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/pion/webrtc/v4"

	grpc_server "github.com/webRTC-broadcasting/grpc"
	"github.com/webRTC-broadcasting/handler"
)

func main(){
	log.SetFlags(0)
	app := fiber.New()

	handler.TrackLocals = map[string]*webrtc.TrackLocalStaticRTP{}

	app.Use("/ws", handler.WsUpgrade)

	app.Get("/ws/:room", websocket.New(handler.WsHandler))
	
	go func() {
        if err := app.Listen(":3000"); err != nil {
            log.Fatalf("failed to start Fiber server: %v", err)
        }
    }()

    grpc_server.StartGRPCServer()
}