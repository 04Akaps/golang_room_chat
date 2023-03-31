package models

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"kakaoTalk/errorHandler"
	"log"
	"net/http"
	"time"
)

var ExistedRoomList []string

const (
	SocketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: SocketBufferSize, WriteBufferSize: messageBufferSize}

type ChatRoomList struct {
	RoomList map[string]*ChatRoom
}

type ChatRoom struct {
	ForwardChannel   chan *Message
	JoinUserChannel  chan *User
	LeaveUserChannel chan *User
	Users            map[*User]bool
	RoomName         string
}

type User struct {
	Socket      *websocket.Conn // client의 웹 소켓
	ChatRoom    *ChatRoom
	Name        string
	SendMessage chan *Message
}

type Message struct {
	Sender  string
	Message string
	To      string
	Time    time.Time
}

type EnterUserReq struct {
	Name     string `json:"name"`
	RoomName string `json:"room_name"`
}

type MakeRoomReq struct {
	RoomName string `json:"room_name"`
}

func NewChatRoom() *ChatRoomList {
	return &ChatRoomList{
		RoomList: make(map[string]*ChatRoom),
	}
}

func (*ChatRoomList) GetRoomList(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(ExistedRoomList)
}

func (roomList *ChatRoomList) MakeRoom(w http.ResponseWriter, r *http.Request) {
	var newRoom MakeRoomReq

	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // body의 최대 크기
	dec := json.NewDecoder(r.Body)                   // decoder
	dec.DisallowUnknownFields()

	err := dec.Decode(&newRoom)

	if err != nil {
		errorHandler.NewHandlerError(w, "잘못된 Body 값", 201)
		return
	}

	if newRoom.RoomName == "" {
		errorHandler.NewHandlerError(w, "잘못된 Body 값", 201)
		return
	}

	for _, room := range ExistedRoomList {
		if room == newRoom.RoomName {
			errorHandler.NewHandlerError(w, "이미 존재하는 방입니다.", 201)
			return
		}
	}

	ExistedRoomList = append(ExistedRoomList, newRoom.RoomName)

	roomList.newRoomCreated(newRoom.RoomName)

	errorHandler.NewHandlerError(w, "방 만들기 성공!", 200)
}

func (room *ChatRoomList) EnterTheRoom(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	var user EnterUserReq

	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // body의 최대 크기
	dec := json.NewDecoder(r.Body)                   // decoder
	dec.DisallowUnknownFields()

	err := dec.Decode(&user)

	if err != nil {
		errorHandler.NewHandlerError(w, "잘못된 Body 값", 201)
		return
	}

	if user.RoomName == "" || user.Name == "" {
		errorHandler.NewHandlerError(w, "잘못된 Body 값", 201)
		return
	}

	Socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("---- serveHTTP:", err)
		return
	}

	enteredRoom := room.RoomList[user.RoomName]

	if enteredRoom == nil {
		errorHandler.NewHandlerError(w, "존재하지 않는 방", 201)
		return
	}

	newUser := &User{
		Socket:      Socket,
		ChatRoom:    enteredRoom,
		Name:        user.Name,
		SendMessage: make(chan *Message, messageBufferSize),
	}

	enteredRoom.JoinUserChannel <- newUser

	defer func() { enteredRoom.LeaveUserChannel <- newUser }()

	go newUser.Write()

	newUser.Read()
}

func (r *ChatRoomList) newRoomCreated(newRoom string) {
	isExisted := false
	for _, room := range ExistedRoomList {
		if room == newRoom {
			isExisted = true
			break
		}
	}

	if !isExisted {
		log.Println("존재하지 않는 방 -> 데이터 오류,  데이터 삭제 진행")
		for i := 0; i < len(ExistedRoomList); i++ {
			if ExistedRoomList[i] == newRoom {
				ExistedRoomList = append(ExistedRoomList[:i], ExistedRoomList[i+1:]...)
				break
			}
		}
		log.Println("데이터 삭제 완료")

		return
	}

	room := r.RoomList[newRoom]
	go room.ListeningMessage()
}

func (r *ChatRoomList) InitRun() {
	for _, roomName := range ExistedRoomList {
		existedRoom := r.RoomList[roomName]
		go existedRoom.ListeningMessage()
	}
}

func (room *ChatRoom) ListeningMessage() {
	for {
		select {
		case newUser := <-room.JoinUserChannel:
			room.Users[newUser] = true

		case leaveUser := <-room.LeaveUserChannel:
			room.Users[leaveUser] = false

		case message := <-room.ForwardChannel:
			for user := range room.Users {
				user.SendMessage <- message
			}

		}

	}
}

func (u *User) Read() {
	// 클라이언트가 ReadMessage메소드를 통해서 소켓에서 읽을 수 있고,
	// 받은 메시지를 room타입에게 계속해서 전송을 한다.
	defer u.Socket.Close()
	for {
		var msg *Message
		err := u.Socket.ReadJSON(&msg)
		if err != nil {
			return
		}

		msg.Time = time.Now()
		msg.Sender = msg.Message

		u.ChatRoom.ForwardChannel <- msg
	}
}

func (u *User) Write() {
	defer u.Socket.Close()
	for msg := range u.SendMessage {
		err := u.Socket.WriteJSON(msg)
		if err != nil {
			return
		}
	}
}
