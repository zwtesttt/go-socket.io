package main

import (
	"fmt"
	socketio "github.com/zwtesttt/go-socket.io"
	"github.com/zwtesttt/go-socket.io/engineio"
	"github.com/zwtesttt/go-socket.io/engineio/transport"
	"github.com/zwtesttt/go-socket.io/engineio/transport/polling"
	"github.com/zwtesttt/go-socket.io/engineio/transport/websocket"
	"log"
	"os"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	uri := "ws://127.0.0.1:8080/api/v1/ws"

	opts := &engineio.Options{
		Transports: []transport.Transport{polling.Default, websocket.Default},
	}
	client, err := socketio.NewClient(uri, opts)
	if err != nil {
		panic(err)
	}

	// Handle an incoming event
	client.OnEvent("reply", func(s socketio.Conn, msg string) {
		log.Println("Receive Message /reply: ", "reply", msg)
	})

	err = client.Connect()
	if err != nil {
		panic(err)
	}

	client.OnDisconnect(func(s socketio.Conn, msg string) {
		log.Println("Disconnected from server:", msg)
		fmt.Println("连接关闭")
		os.Exit(1)
	})

	client.OnError(func(conn socketio.Conn, err error) {
		fmt.Println("Error:", err)
		panic(err)
	})
	// 监听 `ping` 和 `pong` 事件
	client.OnEvent("ping", func(s socketio.Conn) {
		log.Println("Received ping from server")
		client.Emit("pong") // 客户端响应 pong 消息
	})

	client.OnEvent("pong", func(s socketio.Conn) {
		log.Println("Received pong from server")
	})
	client.Emit("reply", "hello")

	// 启动一个goroutine每5秒发送一个ping消息
	handlePingPong(client)

	// Keep the application running
	select {}
}

func handlePingPong(client *socketio.Client) {
	// 计数器
	retries := 0

	// 监听 pong 事件，收到 pong 时将计数器重置为 0
	client.OnEvent("pong", func(s socketio.Conn) {
		log.Println("Received pong from server")
		retries = 0
	})

	// 启动一个 goroutine 每 5 秒发送一个 ping 消息
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				client.Emit("ping")
				log.Println("Sent ping")
				retries++ // 每次发送 ping 消息时增加计数器

				// 如果计数器达到 5 次，则断开连接
				if retries >= 5 {
					log.Println("Failed to receive pong from server after 5 retries. Disconnecting...")
					err := client.Close()
					if err != nil {
						log.Println("Error while closing client connection:", err)
					}
					return
				}
			}
		}
	}()
}
