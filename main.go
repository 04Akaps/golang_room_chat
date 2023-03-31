package main

import (
	"kakaoTalk/server"
	"log"
	"net/http"
)

func main() {

	err := http.ListenAndServe(":80", server.RegisterRouter())

	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if r := recover(); r != nil {
			log.Println("누가 잘못했냐....")
			// 만약 서버가 꺼지게 되면, 처리할 로직 작성 가능
		}
	}()
}
