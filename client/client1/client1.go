package main

import (
	"fmt"
	"net"
)

var Name string = ""
var Password string = ""
var flag bool

func main() {
	//1.建立与服务端的连接
	conn, err := net.Dial("tcp", "127.0.0.1:3000")
	if err != nil {
		fmt.Println("conn fail...")
	}
	defer conn.Close() //关闭连接
	fmt.Println("connect server successed \n")
	fmt.Printf("请问您是要登录还是注册？")
	fmt.Println()
	fmt.Printf("请输入登录或者注册：")
	xiaoxi := ""
	fmt.Scanf("%s", &xiaoxi)
	if xiaoxi == "登录" {
		fmt.Printf("input a nickname:")
		fmt.Scanf("%s", &Name)
		fmt.Printf("input password:")
		fmt.Scanf("%s", &Password)
		conn.Write([]byte("登录Name|" + Name + Password))
		go Handle(conn)
		for {
			var msg string
			msg = ""
			fmt.Scanln(&msg)
			conn.Write([]byte(Name + " say " + msg))
		}

	} else if xiaoxi == "注册" {
		//给自己取一个昵称吧
		fmt.Printf("Make a nickname:")
		fmt.Scanf("%s", &Name)
		fmt.Printf("Make password:")
		fmt.Scanf("%s", &Password)
		//fmt.Printf("恭喜***%s***注册成功",Name)
		fmt.Println()
		conn.Write([]byte("注册Name|" + Name + Password))
		flag = true
		go Handle(conn)
		for {
			var msg string
			msg = ""
			fmt.Scanln(&msg)
			conn.Write([]byte(Name + " say " + msg))
		}
	}

}

func Handle(conn net.Conn) {

	for {
		data := make([]byte, 255)
		msg_read, err := conn.Read(data)
		if msg_read == 0 || err != nil {
			break
		}
		var str string
		str = Name[6:9]
		if msg_read-10 > 0 && msg_read-7 > 0 {
			if (string(data[msg_read-10 : msg_read-7])) != str {
				fmt.Println(string(data[0:msg_read]))
			}
		}else{
			fmt.Println(string(data[0:msg_read]))
		}

	}
}
