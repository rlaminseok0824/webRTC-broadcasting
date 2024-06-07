import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_webrtc/flutter_webrtc.dart';
import 'package:web_socket_client/web_socket_client.dart';

void main() {
  runApp(MyApp());
}

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'Navigation Example',
      theme: ThemeData(
        primarySwatch: Colors.blue,
      ),
      initialRoute: '/',
      routes: {
        '/': (context) => HomePage(),
        '/page1': (context) => Broadcast(),
        '/page2': (context) => EmptyPage(title: 'Page 2'),
      },
    );
  }
}

class HomePage extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('Home'),
        backgroundColor: Colors.blue,
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            ElevatedButton(
              onPressed: () {
                Navigator.pushNamed(context, '/page1');
              },
              child: Text('Broadcasting'),
            ),
            const SizedBox(height: 15),
            ElevatedButton(
              onPressed: () {
                Navigator.pushNamed(context, '/page2');
              },
              child: Text('Viewing'),
            ),
          ],
        ),
      ),
    );
  }
}

class Broadcast extends StatefulWidget {
  const Broadcast({super.key});

  @override
  State<Broadcast> createState() => _BroadcastState();
}

class _BroadcastState extends State<Broadcast> {
  final _localRenderer = RTCVideoRenderer();
  late RTCPeerConnection _peerConnection;
  RTCSessionDescription? remoteDesc;
  bool _isCandidate = false;
  final TextEditingController _controller = TextEditingController();
  late WebSocket _socket;

  @override
  void initState() {
    super.initState();
    init();
  }

  Future<void> initWebRTC() async {
    _peerConnection = await createPeerConnection(
      {
        'iceServers': [
          {'urls': 'stun:stun.l.google.com:19302'},
        ]
      },
      {},
    );

    _peerConnection.onIceCandidate = (event) {
      if (event.candidate != null && !_isCandidate) {
        setState(() {
          _isCandidate = true;
        });
      }
    };

    var localStream = await navigator.mediaDevices
        .getUserMedia({'audio': false, 'video': true});
    _localRenderer.srcObject = localStream;

    localStream.getTracks().forEach((track) {
      _peerConnection.addTrack(track, localStream);
    });

    _peerConnection
        .createOffer()
        .then((offer) => _peerConnection.setLocalDescription(offer));

    await _localRenderer.initialize();
  }

  void initWS() {
    _socket = WebSocket(Uri.parse("ws://localhost:3000/ws/1?isBroadcast=true"));
    _socket.messages.listen((raw) async {
      Map<String, dynamic> data = jsonDecode(raw);
      print("Received data: $data");
      switch (data['event']) {
        case 'track':
          print("Track ID: ${data['data']}");
          break;
        case 'lsp':
          var bytes = base64Decode(data['data']);
          var decoded = utf8.decode(bytes);
          JsonDecoder decoder = const JsonDecoder();
          var sdp = decoder.convert(decoded);
          RTCSessionDescription realSDP =
              RTCSessionDescription(sdp['sdp'], sdp['type']);
          await _peerConnection.setRemoteDescription(realSDP);
          print("Remote Description set ");

          final sendBody =
              const JsonEncoder().convert({'type': "test", 'sdp': "test"});

          var bytes2 = utf8.encode(sendBody);
          var base64Str = base64.encode(bytes2);

          _socket.send(const JsonEncoder()
              .convert({'event': 'track', 'data': base64Str}));
      }
    });
  }

  Future<void> init() async {
    await initWebRTC();
    initWS();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
        appBar: AppBar(
          title: Text('sfu-ws'),
        ),
        body: OrientationBuilder(builder: (context, orientation) {
          return Column(
            children: [
              // TextField(
              //   controller: _controller,
              //   decoration: InputDecoration(
              //     border: OutlineInputBorder(),
              //   ),
              // ),
              // ElevatedButton(
              //     onPressed: () async {
              //       var bytes = base64Decode(_controller.text);
              //       var decoded = utf8.decode(bytes);
              //       JsonDecoder decoder = JsonDecoder();
              //       var sdp = decoder.convert(decoded);
              //       RTCSessionDescription realSDP =
              //           RTCSessionDescription(sdp['sdp'], sdp['type']);
              //       await _peerConnection.setRemoteDescription(realSDP);
              //       print("Remote Description set ");
              //     },
              //     child: Text('Start Broadcast')),
              // const SizedBox(
              //   height: 15,
              // ),
              _isCandidate
                  ? ElevatedButton(
                      onPressed: () async {
                        final lsd = await _peerConnection.getLocalDescription();
                        final sendBody = const JsonEncoder()
                            .convert({'type': lsd!.type, 'sdp': lsd.sdp});

                        var bytes = utf8.encode(sendBody);
                        var base64Str = base64.encode(bytes);
                        print(base64Str);
                        _socket.send(const JsonEncoder().convert({
                          'event': 'offer',
                          'data': base64Str,
                          'id': "video"
                        }));
                      },
                      child: Text('Send Offer'))
                  : const SizedBox(),
              Row(
                children: [
                  Text('Local Video', style: TextStyle(fontSize: 50.0))
                ],
              ),
              Row(
                children: [
                  SizedBox(
                      width: 160,
                      height: 120,
                      child: RTCVideoView(_localRenderer, mirror: true))
                ],
              ),
              Row(
                children: [
                  Text('Remote Video', style: TextStyle(fontSize: 50.0))
                ],
              ),
              Row(
                children: [
                  Text('Logs Video', style: TextStyle(fontSize: 50.0))
                ],
              ),
            ],
          );
        }));
  }
}

///////////////////////////////////////////////////////////

class EmptyPage extends StatefulWidget {
  final String title;

  const EmptyPage({Key? key, required this.title});

  @override
  State<EmptyPage> createState() => _EmptyPageState();
}

class _EmptyPageState extends State<EmptyPage> {
  late RTCPeerConnection _peerConnection;
  late RTCSessionDescription? remoteDesc;
  List _remoteRenderers = [];
  bool _isCandidate = false;
  final TextEditingController _controller = TextEditingController();
  late WebSocket _socket;

  @override
  void initState() {
    super.initState();
    init();
  }

  void initWS() {
    _socket =
        WebSocket(Uri.parse("ws://localhost:3000/ws/2?isBroadcast=false"));
    _socket.messages.listen((raw) async {
      Map<String, dynamic> data = jsonDecode(raw);
      switch (data['event']) {
        case 'lsp':
          var bytes = base64Decode(data['data']);
          var decoded = utf8.decode(bytes);
          JsonDecoder decoder = const JsonDecoder();
          var sdp = decoder.convert(decoded);
          RTCSessionDescription realSDP =
              RTCSessionDescription(sdp['sdp'], sdp['type']);
          await _peerConnection.setRemoteDescription(realSDP);
          print("Remote Description set ");

          final sendBody =
              const JsonEncoder().convert({'type': "test", 'sdp': "test"});

          var bytes2 = utf8.encode(sendBody);
          var base64Str = base64.encode(bytes2);

          _socket.send(const JsonEncoder()
              .convert({'event': 'track', 'data': base64Str}));
      }
    });
  }

  Future<void> initWebRTC() async {
    _peerConnection = await createPeerConnection(
      {
        'iceServers': [
          {'urls': 'stun:stun.l.google.com:19302'},
        ]
      },
      {},
    );

    _peerConnection.onIceCandidate = (event) {
      if (event.candidate != null && !_isCandidate) {
        setState(() {
          _isCandidate = true;
        });
      }
    };

    _peerConnection.addTransceiver(
        kind: RTCRtpMediaType.RTCRtpMediaTypeVideo,
        init: RTCRtpTransceiverInit(
          direction: TransceiverDirection.SendRecv,
        ));

    _peerConnection
        .createOffer()
        .then((offer) => _peerConnection.setLocalDescription(offer));

    _peerConnection.onTrack = (event) async {
      if (event.track.kind == 'video' && event.streams.isNotEmpty) {
        var renderer = RTCVideoRenderer();
        await renderer.initialize();
        renderer.srcObject = event.streams[0];

        setState(() {
          _remoteRenderers.add(renderer);
        });
      }
    };

    _peerConnection.onRemoveStream = (stream) {
      // Set the new renderer list
      setState(() {
        stream.dispose();
        _remoteRenderers = [];
      });
    };
  }

  Future<void> init() async {
    await initWebRTC();
    initWS();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
        appBar: AppBar(
          title: Text('sfu-ws'),
        ),
        body: OrientationBuilder(builder: (context, orientation) {
          return Column(
            children: [
              // TextField(
              //   controller: _controller,
              //   decoration: InputDecoration(
              //     border: OutlineInputBorder(),
              //   ),
              // ),
              // ElevatedButton(
              //     onPressed: () async {
              //       var bytes = base64Decode(_controller.text);
              //       var decoded = utf8.decode(bytes);
              //       JsonDecoder decoder = JsonDecoder();
              //       var sdp = decoder.convert(decoded);
              //       RTCSessionDescription realSDP =
              //           RTCSessionDescription(sdp['sdp'], sdp['type']);
              //       await _peerConnection.setRemoteDescription(realSDP);
              //       print("Remote Description set ");
              //     },
              //     child: Text('Start Broadcast')),
              // const SizedBox(height: 15),
              _isCandidate
                  ? ElevatedButton(
                      onPressed: () async {
                        final lsd = await _peerConnection.getLocalDescription();
                        final sendBody = const JsonEncoder()
                            .convert({'type': lsd!.type, 'sdp': lsd.sdp});

                        var bytes = utf8.encode(sendBody);
                        var base64Str = base64.encode(bytes);
                        print(base64Str);
                        _socket.send(const JsonEncoder()
                            .convert({'event': 'offer', 'data': base64Str}));
                      },
                      child: Text('Start Viewing'))
                  : const SizedBox(),
              Row(
                children: [
                  Text('Remote Video', style: TextStyle(fontSize: 50.0))
                ],
              ),
              Row(
                children: [
                  ..._remoteRenderers.map((remoteRenderer) {
                    return SizedBox(
                        width: 160,
                        height: 120,
                        child: RTCVideoView(remoteRenderer));
                  }).toList(),
                ],
              ),
              Row(
                children: [
                  Text('Logs Video', style: TextStyle(fontSize: 50.0))
                ],
              ),
            ],
          );
        }));
  }

  Future<RTCSessionDescription> sendLocalDescription(
      RTCSessionDescription rsd) async {
    // Send local description to server
    Dio dio = Dio();

    try {
      Response response = await dio.post('http://127.0.0.1:8080/view', data: {
        'sdp': const JsonEncoder().convert({'type': rsd.type, 'sdp': rsd.sdp})
      });

      // Parse response data into RTCSessionDescription
      RTCSessionDescription remoteDescription = RTCSessionDescription(
        response.data['sdp'], // Assuming 'type' is a key in your response data
        response.data['type'], // Assuming 'sdp' is a key in your response data
      );

      return remoteDescription;
    } catch (e) {
      print(e);
      // Return an empty RTCSessionDescription on error
      return RTCSessionDescription('', '');
    }
  }
}

class CustomLogInterceptor extends Interceptor {
  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    print('REQUEST[${options.method}] => PATH: ${options.path}');
    super.onRequest(options, handler);
  }

  @override
  void onResponse(Response response, ResponseInterceptorHandler handler) {
    print(
      'RESPONSE[${response.statusCode}] => PATH: ${response.requestOptions.path}',
    );
    super.onResponse(response, handler);
  }

  @override
  void onError(DioError err, ErrorInterceptorHandler handler) {
    print(
      'ERROR[${err.response?.statusCode}] => PATH: ${err.requestOptions.path}',
    );
    super.onError(err, handler);
  }
}
