package main

import (
	"net"
	"regexp"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	go user.ListenC()
	return user
}

func (user *User) Online() {
	user.server.MapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.MapLock.Unlock()
	user.server.BroadCast(user, "上线了")
}

func (user *User) Offline() {
	user.server.MapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.MapLock.Unlock()
	user.server.BroadCast(user, "下线了")
}

func (user *User) DoMessage(msg string) {
	flag, err := regexp.MatchString("^to\\|[^\\|]+\\|[^\\|]+$", msg)
	msg = strings.ReplaceAll(msg, "\n", "")
	if msg == "who" {
		user.server.MapLock.Lock()
		for _, u := range user.server.OnlineMap {
			online_user := "[" + u.Addr + "]" + u.Name + "在线\n"
			user.sendMsg(online_user)
		}
		user.server.MapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]
		_, ok := user.server.OnlineMap[newName]
		if ok {
			user.sendMsg("当前用户名已被使用")
		} else {
			user.server.MapLock.Lock()
			delete(user.server.OnlineMap, user.Name)
			user.Name = newName
			user.server.OnlineMap[newName] = user
			user.server.MapLock.Unlock()
		}
	} else if err == nil && flag {
		parts := strings.Split(msg, "|")
		name := parts[1]
		send_msg := parts[2]
		user.server.MapLock.Lock()
		u, ok := user.server.OnlineMap[name]
		if ok {
			u.sendMsg("[" + user.Addr + "]" + user.Name + ":" + send_msg + "\n")
		}
		user.server.MapLock.Unlock()
	} else {
		user.server.BroadCast(user, msg)
	}
}

func (user *User) sendMsg(msg string) {
	user.conn.Write([]byte(msg))
}

func (user *User) ListenC() {
	for {
		msg := <-user.C
		user.sendMsg(msg)
	}
}