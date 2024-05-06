package main

import "github.com/pion/webrtc/v4"

func View(recvOnlyOffer webrtc.SessionDescription) {

	// PeerConnection 생성
	peerConnection, err := webrtc.NewPeerConnection(defaultPeerConnectionConfig)
	if err != nil {
		panic(err)
	}

	// RTCP 트랙을 받을 트랙 추가
	rtpSender, err := peerConnection.AddTrack(trackLocals["video"]) //현재는 지정된 id로 localTrack 추가
	if err != nil {
		panic(err)
	}

	err = peerConnection.SetRemoteDescription(recvOnlyOffer)
	if err != nil {
		panic(err)
	}

	answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			panic(err)
		}


	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	<-gatherComplete


	LocalDescriptionChan <- Encode(peerConnection.LocalDescription(), false)
	
	// RTCP 트랙을 받을 채널을 생성하고, RTCP 트랙을 받는다.
	rtcpBuf := make([]byte, 1500)
	for {
		if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
			return
		}
	}
}	