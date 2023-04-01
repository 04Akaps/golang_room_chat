package server

import (
	"fmt"
	"github.com/04Akaps/golang_room_chat/models"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"net/http"
)

func RegisterRouter() http.Handler {

	defer func() {
		fmt.Println("--------- 서버 시작 ------------")
	}()

	router := mux.NewRouter()

	c := cors.AllowAll() // 개발 편의상을 위해 전체 수용
	corsRouter := c.Handler(router)

	socketRouter(router)

	return corsRouter
}

func socketRouter(router *mux.Router) {
	socketRouter := router.PathPrefix("/chat").Subrouter()

	roomList := models.NewChatRoom()

	roomList.InitRun()

	socketRouter.HandleFunc("", roomList.EnterTheRoom)
	socketRouter.HandleFunc("/makeRoom", roomList.MakeRoom).Methods("POST")
	socketRouter.HandleFunc("/roomList", roomList.GetRoomList).Methods("GET")
}
