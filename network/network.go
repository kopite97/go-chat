package network

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"log"
)

type Network struct {
	engin *gin.Engine
}

func NewServer() *Network {
	n := &Network{engin: gin.New()}

	// 미들웨어 설정

	n.engin.Use(gin.Logger()) // Use를 통해 모든API나 라우터에 대해 특정 범용적인 처리를 한다.
	n.engin.Use(gin.Recovery())
	// Recovery는 패닉이 일어나 서버가 죽어버릴때 다시 살리는 역할. 이 2가지는 사용하는게 좋다.
	n.engin.Use(cors.New(cors.Config{
		AllowWebSockets:  true,
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	r := NewRoom()
	go r.RunInit()

	n.engin.GET("/room", r.SocketServe)

	return n
}

func (n *Network) StartServer() error {
	log.Println("Starting Server")
	return n.engin.Run(":8080")
}
