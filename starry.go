package main

import (
	"./util"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
    "bytes"
)

var serverAddress string = "localhost:21024"
var proxyAddress string = "0.0.0.0:21025"

type Client struct {
	Name                  string
	Id                    int
	RemoteAddr, LocalAddr net.Addr
	Conn, ProxyConn       net.Conn
}

func (connection Client) Say(sender, message string) {
	for len(message) > 0 {
		end := 55 - len(sender)
		if len(message) < end {
			end = len(message)
		}
		n_message := message[:end]
		if len(message) >= end {
			message = message[end:]
		}
		length := len(sender) + len(n_message) + 8
		// Normal text (what looks like a player)
		//encoded := []byte{0x05, byte(length * 2), 0x01, 0x00, 0x00, 0x00, 0x00, 0x02}
		encoded := []byte{0x05, byte(length * 2), 0x03, 0x00, 0x00, 0x00, 0x00, 0x00}
		encoded = append(encoded, byte(len(sender)))
		encoded = append(encoded, []byte(sender)...)
		encoded = append(encoded, byte(len(n_message)))
		encoded = append(encoded, []byte(n_message)...)
		encoded = append(encoded, []byte{0x30, 0x04, 0x98, 0x6d, 0x2b, 0x0c, 0x60, 0x04, 0x02, 0x8d, 0xc2, 0x51}...)
		connection.Conn.Write(encoded)
		sender = ""
	}
}
func say(sender, message string) {
	for i := 0; i < len(connections); i++ {
		conn := connections[i]
		//conn.Conn.Write(encoded)
		conn.Say(sender, message)
	}
}
func genMsg(sender, msg string) []byte {
    n_message := msg
    length := len(sender) + len(n_message) + 8
    encoded := []byte{0x05, byte(length * 2), 0x03, 0x00, 0x00, 0x00, 0x00, 0x00}
    encoded = append(encoded, byte(len(sender)))
    encoded = append(encoded, []byte(sender)...)
    encoded = append(encoded, byte(len(n_message)))
    encoded = append(encoded, []byte(n_message)...)
    encoded = append(encoded, []byte{0x30, 0x04, 0x98, 0x6d, 0x2b, 0x0c, 0x60, 0x04, 0x02, 0x8d, 0xc2, 0x51}...)
    return encoded
}
func (connection Client) GiveItem(item string, count int){
    length := len(item) + 4
    encoded := []byte{0x14, byte(length * 2)}
    encoded = append(encoded, byte(len(item)))
    encoded = append(encoded, []byte(item)...)
    encoded = append(encoded, byte( count +1 ))
    encoded = append(encoded, []byte{0x07, 0x00, 0x30, 0x06, 0x83, 0xb7, 0x74, 0x2b, 0x1e, 0x81, 0x1e, 0x0c, 0x02, 0x8d, 0xcb, 0x04, 0x04, 0x8b, 0xa3, 0x4e, 0x08, 0x88, 0x80, 0x01}...)
    connection.Conn.Write(encoded)
}
func giveItem(name, item string, count int) (lines []string) {
	for i := 0; i < len(connections); i++ {
		conn := connections[i]
	    if conn.Name == name {
            lines = append(lines, "Giving "+strconv.Itoa(count)+" "+item+" to "+conn.Name+".")
            conn.GiveItem(item, count)
        }
    }
    return
}
func (connection Client) Console(message string) {
	sender := ""
	for len(message) > 0 {
		end := 55
		if len(message) < end {
			end = len(message)
		}
		n_message := message[:end]
		if len(message) >= end {
			message = message[end:]
		}
		length := len(sender) + len(n_message) + 8
		encoded := []byte{0x05, byte(length * 2), 0x03, 0x00, 0x00, 0x00, 0x00, 0x00}
		encoded = append(encoded, byte(len(sender)))
		encoded = append(encoded, []byte(sender)...)
		encoded = append(encoded, byte(len(n_message)))
		encoded = append(encoded, []byte(n_message)...)
		encoded = append(encoded, []byte{0x30, 0x04, 0x98, 0x6d, 0x2b, 0x0c, 0x60, 0x04, 0x02, 0x8d, 0xc2, 0x51}...)
		connection.Conn.Write(encoded)
	}
}
func broadcast(message string) {
	for i := 0; i < len(connections); i++ {
		conn := connections[i]
		conn.Console(message)
	}
}
func filterConn(dst, src net.Conn) (written int64, err error){
    buf := make([]byte, 32*1024)
    for {
        nr, er := src.Read(buf)
        if nr > 0 {
            //if buf[0] == 0x05 {
            //index := bytes.Index(buf, []byte{0x05, 0x52, 0x03})
            //if index!=-1 {
            if buf[0]==0x05 && buf[2]==0x03 && bytes.Index(buf,[]byte("No such command"))!=-1 {
                length := 16 + int(buf[15])
                //fmt.Println("Dropping message. Length:", length, buf[:length])
                buf = buf[length:]
                nr -= length
            }
            nw, ew := dst.Write(buf[0:nr])
            if nw > 0 {
                written += int64(nw)
            }
            if ew != nil {
                err = ew
                break
            }
            if nr != nw {
                err = io.ErrShortWrite
                break
            }
        }
        if er == io.EOF {
            break
        }
        if er != nil {
            err = er
            break
        }
    }
    return
}
func netProxy(connections chan Client) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", proxyAddress)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	rAddr, err := net.ResolveTCPAddr("tcp", serverAddress)
	if err != nil {
		panic(err)
	}
	for {
		conn, _ := listener.Accept()
		//fmt.Println("Server listerning")
		banned := false
		for i := 0; i < len(bans); i++ {
			ban := bans[i]
			if strings.Index(conn.RemoteAddr().String(), ban.Addr) != -1 {
				fmt.Println("[Info] Banned client tried to connect from IP:", conn.RemoteAddr().String(), "Matched Rule Name:", ban.Name, "Rule IP:", ban.Addr)
				banned = true
				conn.Close()
			}
		}
		if !banned {
			rConn, err := net.DialTCP("tcp", nil, rAddr)
			if err != nil {
				fmt.Println("[Error] Client tried to connect. Server is not accepting connections.")
				conn.Close()
			} else {
				fmt.Println("Remote addr:", rConn.LocalAddr())
				connections <- Client{"Unknown", -1, conn.RemoteAddr(), rConn.LocalAddr(), conn, rConn}
				go filterConn(conn, rConn)
				go io.Copy(rConn, conn)
                //go io.Copy(rConn, conn)
				defer rConn.Close()
			}
			defer conn.Close()
		}
	}
}

type ServerInfo struct {
	Type, Data string
}

var serverPath string = "/home/rice/.starbound/linux64/starbound_server"

func monitorServer(cs chan ServerInfo) {
	for {
		cmd := exec.Command(serverPath)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Println(err)
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			fmt.Println(err)
		}
		err = cmd.Start()
		if err != nil {
			fmt.Println(err)
		}
		_ = stderr
		//go io.Copy(os.Stdout, stdout)
		//go io.Copy(os.Stderr, stderr)
		reader := bufio.NewReader(stdout)
		for {
			line, _, err := reader.ReadLine()
			if err != nil {
				fmt.Println(err)
				break
			} else {
				trim := strings.TrimRight(string(line), "\n")
				if strings.Index(trim, "Info: Client ") == 0 {
					cs <- ServerInfo{"client", trim}
				} else if strings.Index(trim, "Info:  <") == 0 {
					cs <- ServerInfo{"chat", trim}
				} else if strings.Index(trim, "Info: TcpServer") == 0 {
					cs <- ServerInfo{"serverup", trim}
				}
			}
		}
		cmd.Wait()
		fmt.Println("[Error] Server crashed. Rebooting in 3 seconds...")
		time.Sleep(time.Second * 3)
		fmt.Println("[Error] Rebooting...")
	}
}
func printMessages(count int) (lines []string) {
	if count == 0 {
		count = 20
	}
	path := logFile
	if len(path) == 0 {
		path = serverPath
		if path[len(path)-4:] == ".exe" {
			path = path[:len(path)-4]
		}
		path += ".log"
	}

	logs, err := util.ReadLines(path)
	if err != nil {
		lines = append(lines, "[Error] Failed to read log file: "+err.Error()+"Path: "+path)
	}
	if count > len(logs) {
		count = len(logs)
	}
	lines = append(lines, "Last "+strconv.Itoa(count), " log messages (of "+strconv.Itoa(len(lines))+"):")
	for i := 0; i < count; i++ {
		line := len(logs) - count + i
		lines = append(lines, logs[line])
		//fmt.Println(lines[line])
	}
	return
}
func banip(ip, desc string) (lines []string) {
	lines = append(lines, "[Banned] ip: "+ip+" Desc: "+desc)
	bans = append(bans, Ban{ip, desc})
	for i := 0; i < len(connections); i++ {
		conn := connections[i]
		addr_bits := strings.Split(conn.RemoteAddr.String(), ":")
		addr := strings.Join(addr_bits[:len(addr_bits)-1], ":")
		if strings.Index(addr, ip) != -1 {
			lines = append(lines, "[Banned] Name: "+conn.Name+" IP: "+addr)
			broadcast(conn.Name + " has been banned.")
			conn.Conn.Close()
			conn.ProxyConn.Close()
		}
	}
	writeConfig()
	return
}
func processCommand(command string, args []string, ingame bool) (response []string) {
	if ingame {
		for i := 0; i < len(commands); i++ {
			cmd := commands[i]
			if cmd.Command == command && cmd.Auth {
				if len(args) > 0 && args[0] == password {
					args = args[1:]
				} else {
					response = append(response, "Invalid password.")
					response = append(response, printWTF())
					return
				}
			}
		}
	}
	if command == "help" {
        if len(args)==0 {
            response = append(response, genHelp(ingame)...)
        } else {
            for i := 0; i < len(commands); i++ {
                cmd := commands[i]
                if cmd.Command == args[0] {
                    msg := "/" + cmd.Command + " "
                    if cmd.Auth && ingame {
                        msg += "<pass> "
                    }
                    response = append(response, msg+cmd.Fields)
                    response = append(response, "  "+cmd.Description)
                }
            }
        }
	} else if command == "bans" {
		response = append(response, "[Bans]")
		for i := 0; i < len(bans); i++ {
			conn := bans[i]
			response = append(response, "  Name: "+conn.Name+", IP: "+conn.Addr)
		}
	} else if command == "banip" {
		if len(args) > 0 {
			desc := "None"
			if len(args) > 1 {
				desc = strings.Join(args[1:], " ")
			}
			banip(args[0], desc)
		} else {
			response = append(response, "Invalid syntax.")
			response = append(response, printWTF())
		}
	} else if command == "unbanip" {
		if len(args) > 0 {
			//bans = append(bans, Ban{parts[1],"None"})
			for i := 0; i < len(bans); i++ {
				if bans[i].Addr == args[0] {
					conn := bans[i]
					response = append(response, "[Unbanned] Name: ", conn.Name+", IP: "+conn.Addr)
					bans = append(bans[:i], bans[i+1:]...)
					writeConfig()
				} else if strings.Index(bans[i].Addr, args[0]) != -1 {
					response = append(response, "Did you mean: "+bans[i].Name+" instead? (IP: "+bans[i].Addr+")")
				}
			}
		} else {
			response = append(response, "Invalid syntax.")
			response = append(response, printWTF())
		}
	} else if command == "ban" {
		if len(args) > 0 {
			name := strings.Join(args, " ")
			for i := 0; i < len(connections); i++ {
				conn := connections[i]
				addr_bits := strings.Split(conn.RemoteAddr.String(), ":")
				addr := strings.Join(addr_bits[:len(addr_bits)-1], ":")
				if conn.Name == name {
					response = append(response, "[Banned] Name: "+conn.Name+", IP: "+addr)
					bans = append(bans, Ban{addr, conn.Name})
					broadcast(conn.Name + " has been banned.")
					conn.Conn.Close()
					conn.ProxyConn.Close()
					writeConfig()
				} else if strings.Index(conn.Name, name) != -1 {
					response = append(response, "Did you mean: "+conn.Name+" instead? (IP: "+addr+")")
				}
			}
		} else {
			response = append(response, "Invalid syntax.")
			response = append(response, printWTF())
		}
	} else if command == "unban" {
		if len(args) > 0 {
			//bans = append(bans, Ban{parts[1],"None"})
			name := strings.Join(args, " ")
			for i := 0; i < len(bans); i++ {
				if strings.Index(bans[i].Name, name) != -1 {
					conn := bans[i]
					response = append(response, "[Unbanned] Name: "+conn.Name+", IP:"+conn.Addr)
					bans = append(bans[:i], bans[i+1:]...)
					writeConfig()
					break
				}
			}
		} else {
			response = append(response, "Invalid syntax.")
			response = append(response, printWTF())
		}
	} else if command == "kick" {
		if len(args) > 0 {
			//bans = append(bans, Ban{parts[1],"None"})
			name := strings.Join(args, " ")
			for i := 0; i < len(connections); i++ {
				conn := connections[i]
				addr_bits := strings.Split(conn.RemoteAddr.String(), ":")
				addr := strings.Join(addr_bits[:len(addr_bits)-1], ":")
				if conn.Name == name {
					response = append(response, "[Kicked] Name: "+conn.Name+", IP: "+addr)
					broadcast(conn.Name + " has been kicked.")
					conn.Conn.Close()
					conn.ProxyConn.Close()
				} else if strings.Index(conn.Name, name) != -1 {
					response = append(response, "Did you mean: "+conn.Name+" instead? (IP: "+addr+")")
				}
			}
		} else {
			response = append(response, "Invalid syntax.")
			response = append(response, printWTF())
		}
	} else if command == "item" {
        if len(args) == 3 {
            count, _ := strconv.Atoi( args[2] )
            response = append(response, giveItem(args[0], args[1], count )...)
        } else {
			response = append(response, "Invalid syntax.")
			response = append(response, printWTF())
        }
    } else if command == "say" {
		message := strings.Join(args[1:], " ")
		say(args[0], message)
	} else if command == "broadcast" {
		message := strings.Join(args, " ")
		broadcast(message)
	} else if command == "clients" {
		response = append(response, "[Clients]")
		for i := 0; i < len(connections); i++ {
			conn := connections[i]
			response = append(response, "  "+conn.Name+" - ID: "+strconv.Itoa(conn.Id)+", IP: "+conn.RemoteAddr.String())
		}
	} else if command == "log" {
		count := 20
		if len(args) == 1 {
			count, _ = strconv.Atoi(args[0])
		}
		response = append(response, printMessages(count)...)
	} else {
		if len(command) > 0 {
			response = append(response, "Unknown command: "+command)
		}
		response = append(response, printWTF())
	}
	return
}
func cli() {
	fmt.Println("Starry CLI - Welcome to Starry!")
	fmt.Println("Starry is a command line Starbound and remote access administration tool.")
	printWTF()
	for {
		fmt.Print("> ")
		reader := bufio.NewReader(os.Stdin)
		raw_input, _ := reader.ReadString('\n')
		trimmed := strings.TrimRight(raw_input, "\n")
		parts := strings.Split(trimmed, " ")
		command := parts[0]
		resp := processCommand(command, parts[1:], false)
		for i := 0; i < len(resp); i++ {
			fmt.Println(resp[i])
		}
	}
}

type Command struct {
	Command, Fields, Description, Category string
	Auth                                   bool
}

var commands []Command

func genHelp(ingame bool) (lines []string) {
	lines = append(lines, "[Commands]")
	categories := make([]string, 0)
	for i := 0; i < len(commands); i++ {
		command := commands[i]
		found := false
		for j := 0; j < len(categories); j++ {
			if categories[j] == command.Category {
				found = true
				break
			}
		}
		if !found {
			categories = append(categories, command.Category)
		}
	}
	for i := 0; i < len(categories); i++ {
		category := categories[i]
		lines = append(lines, category+":")
		for i := 0; i < len(commands); i++ {
			command := commands[i]
            if command.Category == category {
                msg := "  "
                if ingame {
                    msg += "/"
                }
                msg += command.Command + " "
                if command.Auth && ingame {
                    msg += "<pass> "
                }
                lines = append(lines, msg+command.Fields)
                if !ingame {
                    lines = append(lines, "    "+command.Description)
                }
            }
		}
	}
	return
}
func printWTF() string {
	return "Type 'help' for more information."
}

type Ban struct {
	Addr, Name string
}

var bans []Ban
var connections []Client

var logFile string

type Config struct {
	ServerPath    string
	LogFile       string
	ServerAddress string
	ProxyAddress  string
	Password      string
	Bans          []Ban
}

var password string = "changethis"

func writeConfig() {
	config := Config{serverPath, logFile, serverAddress, proxyAddress, password, bans}
	b, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		fmt.Println("[Error] Failed to create JSON config.")
	}
	util.WriteLines([]string{string(b)}, "starry.config")
}
func readConfig() {
	lines, err := util.ReadLines("starry.config")
	if err != nil {
		fmt.Println("[Error]", err)
		writeConfig()
	} else {
		var config Config
		err := json.Unmarshal([]byte(strings.Join(lines, "\n")), &config)
		if err != nil {
			fmt.Println("[Error]", err)
		} else {
			serverPath = config.ServerPath
			logFile = config.LogFile
			serverAddress = config.ServerAddress
			proxyAddress = config.ProxyAddress
			password = config.Password
			bans = config.Bans
		}
	}
}

func main() {
	commands = []Command{
		Command{"clients", "", "Display connected clients.", "General", false},
		Command{"say", "<sender name> <message>", "Say something.", "General", true},
		Command{"broadcast", "<message>", "Show grey text in chat.", "General", true},
		Command{"help", "[<command>]", "Information on commands.", "General", false},
		Command{"log", "[<count>]", "Last <count> or 20 log messages.", "General", true},
		Command{"nick", "<name>", "Change your character's name. In game only.", "General", false},
		Command{"item", "<name> <item> <count>", "Give items to a player", "General", true},
		Command{"bans", "", "Show ban list.", "Bans", true},
		Command{"ban", "<name>", "Ban an IP by player name.", "Bans", true},
		Command{"banip", "<ip> [<name>]", "Ban an IP or range (eg. 8.8.8.).", "Bans", true},
		Command{"unban", "<name>", "Unban an IP by name.", "Bans", true},
		Command{"unbanip", "<ip>", "Unban an IP.", "Bans", true},
	}
	readConfig()
	clientChan := make(chan Client)
	go netProxy(clientChan)
	serverChan := make(chan ServerInfo)
	go monitorServer(serverChan)
	go cli()
	for {
		select {
		case info, ok := <-serverChan:
			if !ok {
				fmt.Println("Server Monitor channel closed!")
			} else {
				if info.Type == "client" {
					// Info: Client 'Tom' <1> (127.0.0.1:52029) connected
					// Info: Client 'Tom' <1> (127.0.0.1:52029) disconnected
					parts := strings.Split(info.Data, " ")
					op := parts[len(parts)-1]
					if op == "connected" {
						path := parts[len(parts)-2]
						path = path[1 : len(path)-1]
						id_str := parts[len(parts)-3]
						id_str = id_str[1 : len(id_str)-1]
						id, _ := strconv.Atoi(id_str)
						name := strings.Join(parts[2:len(parts)-3], " ")
						name = name[1 : len(name)-1]
						//fmt.Println("Path", path, "Name", name)
						for i := 0; i < len(connections); i++ {
							addr := connections[i].LocalAddr.String()
							//fmt.Println("NPath:", addr, "Path:", path, addr == path)
							if addr == path {
								connections[i].Name = name
								connections[i].Id = id
								fmt.Println("[Client]", connections[i])
								broadcast(connections[i].Name + " has joined.")
							}
						}
					} else if op == "disconnected" {
						id_str := parts[len(parts)-3]
						id_str = id_str[1 : len(id_str)-1]
						id, _ := strconv.Atoi(id_str)
						for i := 0; i < len(connections); i++ {
							if connections[i].Id == id {
								broadcast(connections[i].Name + " has left.")
								connections = append(connections[:i], connections[i+1:]...)
								break
							}
						}

					}
				} else if info.Type == "serverup" {
					fmt.Println("Server listening for connections.")
				} else if info.Type == "chat" {
					parts := strings.Split(info.Data, " ")
					user := parts[2]
					user = user[1 : len(user)-1]
					message := strings.Join(parts[3:], " ")
					fmt.Println("<" + user + "> " + message)
					if string(message[0]) == "/" {
						command := parts[3][1:len(parts[3])]
						if command == "nick" {
						} else {
							resp := processCommand(command, parts[4:], true)
							for j := 0; j < len(connections); j++ {
								conn := connections[j]
								if conn.Name == user {
									for i := 0; i < len(resp); i++ {
										conn.Console(resp[i])
										//fmt.Println(resp[i])
									}
								}
							}
						}
					}
				} else {
					fmt.Println("[ServerMon] Info:", info.Type, "Data:", info.Data)
				}
			}
		case client, ok := <-clientChan:
			if !ok {
				fmt.Println("Server Monitor channel closed!")
			} else {
				connections = append(connections, client)
			}
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
	}
}
