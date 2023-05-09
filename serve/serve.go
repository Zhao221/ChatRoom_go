package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"

	"fmt"
	"net"
	"strings"
	"time"
)

var db *sql.DB // 是一个连接池对象 db全局变量
var flag bool
var id int
var password string

func main() {
	// 数据库相关代码
	err := initDB()
	if err != nil {
		fmt.Printf("init DB failed,err:%v\n", err)
	}
	fmt.Println("数据库连接成功")
	// 聊天室相关代码***************************************************************************
	// 创建监听套接字(socket) 1.开启服务
	listener, err := net.Listen("tcp", "127.0.0.1:3000")
	if err != nil {
		fmt.Println("server start error")
		return
	}

	defer listener.Close()

	fmt.Println("server is wating ....")
	// 创建管理者go程，管理map和全局channel
	go Manager()
	// 循环监听客户端连接请求
	for {
		// 2.等待客户端建立连接
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("conn fail ...", err)
			return
		}
		fmt.Println(conn.RemoteAddr(), "connect successed")
		// 3.启动go程处理客户端数据请求
		go handle(conn)
	}
}

// 数据库函数
func initDB() (err error) {
	// DSN:Data Source Name
	// 数据库信息
	dsn := "root:123456@tcp(127.0.0.1:3306)/Clients"
	// 连接数据库
	db, err = sql.Open("mysql", dsn) // open不会校验用户名和密码是否对不对
	if err != nil {
		return
	}
	err = db.Ping() // 尝试连接数据库，并校验用户名和密码是否正确
	if err != nil {
		return
	}
	// 设置数据库连接池的最大连接数
	// db.SetMaxIdleConns(10)
	// 设置最大空闲连接数
	// db.SetMaxIdleConns(5)
	return
}

// Client 创建用户结构体类型
type Client struct {
	C    chan string
	Name string
	Addr string
}

// 创建全局map，存储在线用户
var onlineMap map[string]Client

// 创建全局 channel 传递用户信息
var message = make(chan string)

func WriteMsgToClient(clnt Client, conn net.Conn) {
	// 监听用户自带channel上是否有消息。
	for msg := range clnt.C {
		str := strings.Contains(msg, "Name")
		if str == true {
			password = msg[20:len(msg)]
			var con string
			con = msg[11:20]
			queryMore(0, con, password)
			if id == 0 {
				conn.Write([]byte(msg + "\n"))
			} else {
				msg = "此账号已注册请重新注册"
			}

		} else {
			conn.Write([]byte(msg + "\n"))
		}
	}
}
func MakeMsg(clnt Client, msg string) (buf string) {
	// fmt.Printf(msg)
	if msg[0:6] == "注册" {
		fmt.Println(msg[0:6] + "2")
		password = msg[20:len(msg)]
		var con string
		con = msg[11:20]
		queryMore(0, con, password)
		fmt.Println(id)
		if id != 0 {
			buf = "[" + clnt.Addr + "]" + clnt.Name + ": " + msg
		} else {
			buf = "什么登陆成功，骗你的，哈哈哈\n此用户已经注册，请重新***连接***注册"
		}
	} else if msg[0:6] == "登录" {
		buf = "[" + clnt.Addr + "]" + clnt.Name + ": " + msg
	} else {
		buf = "[" + clnt.Addr + "]" + clnt.Name + ": " + msg
	}

	return
}
func handle(conn net.Conn) {
	defer conn.Close() // 处理完之后，关闭这个连接
	// 创建channel判断用户是否活跃
	hasData := make(chan bool)
	// 获取用户 网络地址 IP+port
	netAddr := conn.RemoteAddr().String()
	// 创建新连接用户的结构体,默认用户名是IP+port
	clnt := Client{make(chan string), netAddr, netAddr}
	// 将新连接用户，添加到在线用户map中，key：IP+port  value：client2
	onlineMap[netAddr] = clnt

	// 创建专门用来给当前用户发送消息的go程
	go WriteMsgToClient(clnt, conn)
	// 发送用户上线消息到全局channel中
	// message <- "[" + netAddr + "]" + clnt.Name + "登录"
	message <- MakeMsg(clnt, "登录成功")
	// 创建一个channel，用来判断退出状态
	isQuit := make(chan bool)

	// 创建一个匿名go程，专门处理用户发送的消息。
	go func() {
		buf := make([]byte, 4096)

		for {
			n, err := conn.Read(buf)
			// 将读到的用户消息，保存到msg中,string类型
			msg := string(buf[:n])
			if n == 0 {
				isQuit <- true
				fmt.Printf("检测到客户端:%s退出\n", clnt.Name)
				return
			}
			if err != nil {
				fmt.Println("conn.Read err:", err)
				return
			}
			fmt.Println(msg[0:6] + "3")
			msg_str := strings.Split(string(buf[0:n]), "|")
			str := strings.Contains(string(buf[0:n]), "Name")
			var con string

			if str == true {
				if msg[0:6] == "注册" {
					con = msg[11:20]
					password = msg[20:len(msg)]
					queryMore(0, con, password)
					if id == 0 {
						sqlStr := "insert into user(name, password) values (?,?)"
						stmt, err := db.Prepare(sqlStr)
						if err != nil {
							fmt.Printf("prepare failed, err:%v\n", err)
							return
						}
						defer stmt.Close()
						_, err = stmt.Exec(con, password)
						if err != nil {
							fmt.Printf("insert failed, err:%v\n", err)
							return
						}
						fmt.Printf("%s登录成功", msg_str)
						clnt.Name = con
						// 名字更新并保存到onlineMap中
						onlineMap[netAddr] = clnt
					} else {
						fmt.Println(clnt.Name + "：此用户已经注册，请重新注册")
						conn.Write([]byte("什么登陆成功，骗你的，哈哈哈\n此用户已经注册，请重新***连接***注册\n"))
					}
				} else if msg[0:6] == "登录" {
					queryMore(0, con, password)
					if id == 0 {
						conn.Write([]byte("什么登陆成功，骗你的，哈哈哈\n请您先***连接***注册***再***登录\n"))
						return
					}
					count := 0
					con = msg[11:20]
					for _, user := range onlineMap {
						if con == user.Name {
							fmt.Println(clnt.Name + "：此用户已经登录，请重新登录")
							conn.Write([]byte("什么登陆成功，骗你的，哈哈哈\n此用户已经登录，请重新***连接***登录\n"))
						}
					}
					for _, user := range onlineMap {
						if con != user.Name {
							count++
						}
					}
					if count == len(onlineMap) {
						fmt.Printf("%s登录成功", msg_str)
						con = msg[11:20]
						clnt.Name = con
						// 名字更新并保存到onlineMap中
						onlineMap[netAddr] = clnt
					}
				} else {
					fmt.Printf("%s登录成功", msg_str)
					con = msg[11:20]
					clnt.Name = con
					// 名字更新并保存到onlineMap中
					onlineMap[netAddr] = clnt
				}
			}

			if str == false {
				var rename string
				for i := 1; i < len(msg_str); i++ {
					rename += msg_str[i]
				}

				if clnt.Name != rename && rename != "" {
					fmt.Printf("%s:%s改名：%s", clnt.Addr, clnt.Name, rename)
				} else {
					fmt.Println(msg_str)
				}
			}
			fmt.Println()
			content1 := msg[len(msg)-3 : len(msg)]
			// 提取在线用户列表
			if content1 == "who" && len(content1) == 3 {
				conn.Write([]byte("online user list:\n"))
				// 遍历当前map，获取在线用户
				for _, user := range onlineMap {
					userInfo := user.Addr + ":" + user.Name + "\n"
					conn.Write([]byte(userInfo))
				}
				// 判断用户发送了改名命令
			} else if len(msg) >= 22 && msg[14:20] == "rename" { // rename/
				newName := strings.Split(msg, "|")[1]     // msg[8:]
				clnt.Name = newName                       // 修改结构体成员name
				onlineMap[netAddr] = clnt                 // 更新onlineMap
				conn.Write([]byte("rename successful\n")) // 告知用户更新成功
			} else if len(msg) >= 22 && msg[11:17] == "rename" {
				newName := strings.Split(msg, "|")[1]     // msg[8:]
				clnt.Name = newName                       // 修改结构体成员name
				onlineMap[netAddr] = clnt                 // 更新onlineMap
				conn.Write([]byte("rename successful\n")) // 告知用户更新成功
			} else {
				// 私聊的消息格式：[对方IP]消息内容
				// 群发消息：消息内容
				// 通过正则表达式获取对方IP地址
				var reg string
				var count int
				if len(msg) >= 26 {
					if msg[0:9] != msg[17:26] && msg[14:17] == "to@" {
						reg = msg[17:26]
						// 遍历当前map，获取在线用户
						for _, user := range onlineMap {
							if reg == user.Name {
								// 私聊
								// 如果找到Client，那么就给该Client的通道写入消息
								userInfo := "[" + clnt.Addr + "]" + clnt.Name + ":" + msg + "\n"
								user.C <- userInfo
							}
						}
						for _, user := range onlineMap {
							if reg != user.Name {
								count++
							}
						}
						if count == len(onlineMap) {
							fmt.Printf("%s查无此人", reg)
							var reg2 string
							reg2 = msg[0:9]
							for _, user := range onlineMap {
								if reg2 == user.Name {
									// 如果找到Client，那么就给该Client的通道写入消息
									userInfo := "查无此人"
									user.C <- reg + userInfo
								}
							}

						}

					} else {
						// 群发消息
						// 将读到的用户信息，写到message中。
						message <- MakeMsg(clnt, msg)
					}
				} else {
					// 群发消息
					// 将读到的用户信息，写到message中。
					message <- MakeMsg(clnt, msg)
				}

			}
			hasData <- true
		}
	}()
	// 保证 不退出
	for {
		// 监听channel上的数据流动
		select {
		case <-isQuit:
			delete(onlineMap, clnt.Addr)   // 将用户从online移除
			message <- MakeMsg(clnt, "退出") // 写入用户退出消息到全局channel
			return
		case <-hasData:
			// 什么都不做。目的是重置下面case的计时器
		case <-time.After(time.Second * 3600):
			fmt.Printf("%s:%s:超时退出", clnt.Addr, clnt.Name)
			delete(onlineMap, clnt.Addr)     // 将用户从online移除
			message <- MakeMsg(clnt, "超时退出") // 写入用户退出消息到全局channel
			return
		}
	}
}

type user struct {
	id       int
	name     string
	password string
}

func queryMore(n int, name string, password string) {
	// 1. SQL语句
	sqlStr := `select id ,name,password from user where id>?;`
	// 2.执行
	rows, err := db.Query(sqlStr, n)
	if err != nil {
		fmt.Printf("exec %s query failed,err:%v\n", sqlStr, err)
		return
	}
	// 3.一定要关闭rows
	defer rows.Close()
	// 4.循环取值
	for rows.Next() {
		var u1 user
		err := rows.Scan(&u1.id, &u1.name, &u1.password)
		if err != nil {
			fmt.Printf("Scan failed,err:%v\n", err)
		}
		if name == u1.name && password == u1.password {
			id = u1.id
			return
		} else {
			id = 0
		}
	}
}
func Manager() {
	// 初始化onlineMap
	onlineMap = make(map[string]Client)

	for {
		// 监听全局channel中是否有数据，有数据存储至msg，无数据阻塞
		msg := <-message
		// 循环发送消息给所有在线用户 。要想执行，必须msg:=<-message执行完，解除阻塞。
		for _, clnt := range onlineMap {

			clnt.C <- msg
		}
	}
}
