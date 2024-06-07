# WebRTC-Broadcasting
---
Thanks to Fiber & pion webRTC, I make a broadcasting webRTC service on Fiber websocket. 

- Reference
	+  https://github.com/pion/example-webrtc-applications/tree/master/sfu-ws
	+  https://github.com/pion/webrtc/tree/master/examples/broadcast



# Getting Start

## 1. Clone Repository
``` bash
git clone https://github.com/rlaminseok0824/webRTC-broadcasting.git
cd webRTC-broadcasting
```

## 2. Install Dependencies

``` bash
go mod tidy
```
## 3. Run Service

``` bash
go run main.go
```

# Quick Start

1. you can start with Dockerfile

``` bash
 docker build -t webrtc-broadcasting .
```

2. Then, start docker with below code.

``` bash
docker run -p 3000:3000 -p 4040:4040 webrtc-broadcasting
```


# Project Structure
```
📦webRTC-broadcasting  
  ┣ 📂grpc  
 ┃ ┗ 📜server.go  
 ┣ 📂handler  
 ┃ ┣ 📜broadcast.go  
 ┃ ┣ 📜handler.go  
 ┃ ┣ 📜model.go  
 ┃ ┗ 📜view.go  
 ┣ 📂proto  
 ┃ ┣ 📜service.pb.go  
 ┃ ┣ 📜service.proto  
 ┃ ┗ 📜service_grpc.pb.go  
 ┣ 📂utils  
 ┃ ┗ 📜convert.go  
 ┣ 📜.gitignore  
 ┣ 📜Dockerfile  
 ┣ 📜README.md  
 ┣ 📜go.mod  
 ┣ 📜go.sum  
 ┗ 📜main.go
```
