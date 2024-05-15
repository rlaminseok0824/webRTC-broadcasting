package main

import (
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/webrtc/v4"
)


func Broadcast(offer webrtc.SessionDescription){
	// webRTC의 정보를 보내는 broadcast peerConnection 생성
	
	// webRTC의 기본 미디어 코덱 엔진 등록
	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		panic(err)
	}

	// webRTC 에서 RTP/RTCP 파이프라인 생성 
	// NACKS, RTCP Reports 등을 처리하기 위해 사용됨
	// 현재는 default로 생성 => NewPeerConnection 만들면 자동 default로 생성됨
	i := &interceptor.Registry{}

	if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
		panic(err)
	}


	//interceptor 추가
	//매 3초마다 RTCP 패킷(video keyframe)을 보낸다.
	//realworld에서는 직접 RTCP 패킷을 만들어서 보내야함. -> Future Plan
	intervalPliFactory, err := intervalpli.NewReceiverInterceptor()
	if err != nil {
		panic(err)
	}
	i.Add(intervalPliFactory)


	//webRTC peerConnection 생성
	peerConnection, err := webrtc.NewAPI(webrtc.WithMediaEngine(m),webrtc.WithInterceptorRegistry(i)).NewPeerConnection(defaultPeerConnectionConfig)
	if err != nil {
		panic(err)
	}
	defer func() {
		if cErr := peerConnection.Close(); cErr != nil {
			fmt.Printf("cannot close peerConnection: %v\n", cErr)
		}
	}()
	
	//Video + Audio Track 하나만을 받음
	for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		if _, err := peerConnection.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		}); err != nil {
			log.Print(err)
			return
		}
	}

	//RemoteTrack이 들어오면 실행되는 함수
	peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) { 
		// Create a local track, all our SFU clients will be fed via this track
		localTrack, newTrackErr := webrtc.NewTrackLocalStaticRTP(remoteTrack.Codec().RTPCodecCapability, "video", "pion") //"video","pion"은 추후 수정 필요
		if newTrackErr != nil {
			panic(newTrackErr)
		}

		// thread 상에서 작동하기 때문에 channel을 통해 localTrack을 전달
		localTrackChan <- localTrack
		
		//rtp Buffer 생성해서 전달
		rtpBuf := make([]byte, 1500)
		for {
			i, _, readErr := remoteTrack.Read(rtpBuf)
			if readErr != nil {
				panic(readErr)
			}

			// ErrClosedPipe means we don't have any subscribers, this is ok if no peers have connected yet
			if _, err = localTrack.Write(rtpBuf[:i]); err != nil && !errors.Is(err, io.ErrClosedPipe) {
				panic(err)
			}
		}
	})

	// Offer로 받은 정보를 remoteDescription으로 설정
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	// 이를 바탕으로 answer 생성
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// ICE 서버를 자동으로 도와주는 함수
	// Channel를 통해 ICE Candidate를 전달 받음
	// realworld에선 직접 ICE Candidate 호출 -> Future Plan
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	log.Println("Gathering ICE Candidates")
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	//ICE Candidate 정보 끝날때까지 기다림
	<-gatherComplete

	log.Println("Gathering ICE Candidates Finished")
	//peerConnection의 LocalDescription을 전달
	LocalDescriptionChan <- Encode(peerConnection.LocalDescription(),false)

	//함수가 죽지 않게 대기
	select{}
}