package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"sync"
	"math"
	//"reflect"
	//"encoding/binary"

	"github.com/gorilla/websocket"

	"github.com/pion/webrtc/v3"
	"github.com/wawesomeNOGUI/webrtcGamerServer/internal/signal"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade: ", err)
		return
	}
	defer c.Close()
	fmt.Println("User connected from: ", c.RemoteAddr())

	//===========This Player's Variables===================
	var playerTag string

	//===========WEBRTC====================================
	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	//Setup dataChannel to act like UDP with ordered messages (no retransmits)
	//with the DataChannelInit struct
	var udpPls webrtc.DataChannelInit
	var retransmits uint16 = 0

	//DataChannel will drop any messages older than
	//the most recent one received if ordered = true && retransmits = 0
	//This is nice so we can always assume client
	//side that the message received from the server
	//is the most recent update, and not have to
	//implement logic for handling old messages
	var ordered = true

	udpPls.Ordered = &ordered
	udpPls.MaxRetransmits = &retransmits

	// Create a datachannel with label 'UDP' and options udpPls
	dataChannel, err := peerConnection.CreateDataChannel("UDP", &udpPls)
	if err != nil {
		panic(err)
	}

	//Create a reliable datachannel with label "TCP" for all other communications
	reliableChannel, err := peerConnection.CreateDataChannel("TCP", nil)
	if err != nil {
		panic(err)
	}

	// Set the handler for ICE connection state

	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State has changed: %s\n", connectionState.String())

		//3 = ICEConnectionStateConnected
		if connectionState == 3 {
			//Store a new x and y for this player
			NumberOfPlayers++
			playerTag = strconv.Itoa(NumberOfPlayers)
			fmt.Println(playerTag)

			//Store a slice for player x, y, and other data
			//Initially we'll just have starting x, y, and x speed values
			Updates.Store(playerTag, []int{0, 0, 0})

		}else if connectionState == 5 || connectionState == 6 || connectionState == 7{
			Updates.Delete(playerTag)
			fmt.Println("Deleted Player")
		}
	})

//====================No retransmits, ordered dataChannel=======================
	// Register channel opening handling
	dataChannel.OnOpen(func() {
		fmt.Printf("Data channel '%s'-'%d' open. Random messages will now be sent to any connected DataChannels\n", dataChannel.Label(), dataChannel.ID())

		for {
			time.Sleep(time.Millisecond*50) //50 milliseconds = 20 updates per second
			                                //20 milliseconds = ~60 updates per second


			//fmt.Println(UpdatesString)
			// Send the message as text so we can JSON.parse in javascript
			sendErr := dataChannel.SendText(UpdatesString)
			if sendErr != nil {
				fmt.Println("data send err", sendErr)
				break
			}
		}

	})

	// Register text message handling
	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		fmt.Printf("Message from DataChannel '%s': '%s'\n", dataChannel.Label(), string(msg.Data))
	})

//==============================================================================

//=========================Reliable DataChannel=================================
	// Register channel opening handling
	reliableChannel.OnOpen(func() {

			//Send Client their playerTag so they know who they are in the Updates Array
			sendErr := reliableChannel.SendText(playerTag)
			if sendErr != nil {
				panic(err)
			}


	})

	// Register message handling (Data all served as a bytes slice []byte)
	// for user controls
	reliableChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		//fmt.Printf("Message from DataChannel '%s': '%s'\n", reliableChannel.Label(), string(msg.Data))

		//fmt.Println(msg.Data)
		if msg.Data[0] == 88 {   //88 = "X"
			x, err := strconv.Atoi( string(msg.Data[1:]) );
				if err != nil{
					fmt.Println(err)
				}

				playerSlice, ok := Updates.Load(playerTag)
				if ok == false {
					fmt.Println("Uh oh")
				}

				playerSlice.([]int)[0] = x
				Updates.Store( playerTag, playerSlice )
		}else if msg.Data[0] == 89 {  //89 = "Y"
		 y, err := strconv.Atoi( string(msg.Data[1:]) );
				if err != nil{
					fmt.Println(err)
				}

			playerSlice, ok := Updates.Load(playerTag)
			if ok == false {
				fmt.Println("Uh oh")
			}

			playerSlice.([]int)[1] = y
			Updates.Store( playerTag, playerSlice )
		}else if msg.Data[0] == 83 {  //83 = "S"
		 s, err := strconv.Atoi( string(msg.Data[1:]) );
				if err != nil{
					fmt.Println(err)
				}

			playerSlice, ok := Updates.Load(playerTag)
			if ok == false {
				fmt.Println("Uh oh")
			}

			playerSlice.([]int)[2] = s
			Updates.Store( playerTag, playerSlice )
		}
	})

//==============================================================================

	// Create an offer to send to the browser
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(offer)
	if err != nil {
		panic(err)
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	fmt.Println(*peerConnection.LocalDescription())

	//Send the SDP with the final ICE candidate to the browser as our offer
	err = c.WriteMessage(1, []byte(signal.Encode(*peerConnection.LocalDescription()))) //write message back to browser, 1 means message in byte format?
	if err != nil {
		fmt.Println("write:", err)
	}

	//Wait for the browser to send an answer (its SDP)
	msgType, message, err2 := c.ReadMessage() //ReadMessage blocks until message received
	if err2 != nil {
		fmt.Println("read:", err)
	}

	answer := webrtc.SessionDescription{}

	signal.Decode(string(message), &answer) //set answer to the decoded SDP
	fmt.Println(answer, msgType)

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(answer)
	if err != nil {
		panic(err)
	}

	//=====================Trickle ICE==============================================
	//Make a new struct to use for trickle ICE candidates
	var trickleCandidate webrtc.ICECandidateInit
	var leftBracket uint8 = 123 //123 = ascii value of "{"

	for {
		_, message, err2 := c.ReadMessage() //ReadMessage blocks until message received
		if err2 != nil {
			fmt.Println("read:", err)
		}

		//If staement to make sure we aren't adding websocket error messages to ICE
		if message[0] == leftBracket {
			//Take []byte and turn it into a struct of type webrtc.ICECandidateInit
			//(declared above as trickleCandidate)
			err := json.Unmarshal(message, &trickleCandidate)
			if err != nil {
				fmt.Println("errorUnmarshal:", err)
			}

			fmt.Println(trickleCandidate)

			err = peerConnection.AddICECandidate(trickleCandidate)
			if err != nil {
				fmt.Println("errorAddICE:", err)
			}
		}

	}

}

//We'll have this marshalling function here so the multiple gorutines for each
//player will not be inefficient by all trying to marshall the same thing
func getSyncMapReadyForSending(m *sync.Map){
	for{
		time.Sleep(time.Millisecond)

		tmpMap := make(map[string][]int)
    m.Range(func(k, v interface{}) bool {
        tmpMap[k.(string)] = v.([]int)
        return true
    })

    jsonTemp, err := json.Marshal(tmpMap)
		if err != nil{
			panic(err)
		}

		UpdatesString = string(jsonTemp)
	}
}


func gameSimulation(m *sync.Map){
	for{
		m.Range(func (key, value interface{}) bool {

			if math.Abs(value.([]int)[0] + 25 - ball[0]) < 25 && math.Abs(value.([]int)[1] + 12 - ball[1]) < 12 {
				ball[1] = value.([]int)[1] + 5  //move ball up a bit for bounce
				ball[3] = ball[3]*-1    //switch y speed
				ball[2] = value.([]int)[2] / 2  //set x speed
			}

			//If our function passed to Range returns false, Range stops iteration
			return true
		})
	}
}


//Updates Map
var Updates sync.Map
var UpdatesString string

//Ball Info Slice
//index 0 = x, 1 = y, 2 = xSpeed, 3 = ySpeed
var ball = []int{0, 0, 0, 0}

var NumberOfPlayers int


func main() {

	go getSyncMapReadyForSending(&Updates)
	go gameSimulation(&Updates)

	fileServer := http.FileServer(http.Dir("./public"))
	http.HandleFunc("/echo", echo) //this request comes from webrtc.html
	http.Handle("/", fileServer)

	err := http.ListenAndServe(":80", nil) //Http server blocks
	if err != nil {
		panic(err)
	}
}
