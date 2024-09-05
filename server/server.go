package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	MapLock   sync.RWMutex
	Message   chan string
}

func NewServer(Ip string, Port int) *Server {
	return &Server{
		Ip:        Ip,
		Port:      Port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}
func (s *Server) handle(conn net.Conn) {
	user := NewUser(conn, s)
	user.Online()
	isLive := make(chan bool)
	go func() {
		reader := bufio.NewReader(conn)
		for {
			message, err := reader.ReadString('\n')
			if err != nil {
				user.Offline()
				break
			}
			isLive <- true
			user.DoMessage(message)
		}
	}()
	for {
		select {
		case <-isLive:
		case <-time.After(1000 * time.Second):
			user.sendMsg("您已超时，将被强制下线")
			removeMsg := "[" + user.Addr + "]" + user.Name + "超时，系统将其强制下线\n"
			fmt.Println(removeMsg)
			close(user.C)
			conn.Close()
			return
		}
	}
}

func (s *Server) BroadCast(user *User, msg string) {
	bc_msg := "[" + user.Addr + "]" + user.Name + ":" + msg + "\n"
	s.Message <- bc_msg
}

func (s *Server) ListenMessager() {
	for {
		msg := <-s.Message
		s.MapLock.Lock()
		for _, user := range s.OnlineMap {
			user.C <- msg
		}
		s.MapLock.Unlock()
	}
}

func (s *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer listener.Close()

	go s.ListenMessager()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}
		go s.handle(conn)
	}
}

func main() {
	s := NewServer("127.0.0.1", 8888)
	s.Start()
}