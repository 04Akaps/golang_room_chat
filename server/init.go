package server

import (
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"kakaoTalk/models"
	"net/http"
)

func RegisterRouter() http.Handler {
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
