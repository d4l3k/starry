package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var remoteAddr string = "localhost:21024"

type Client struct {
	Name                  string
	Id                    int
	RemoteAddr, LocalAddr net.Addr
	Conn, ProxyConn       net.Conn
}

func netProxy(connections chan Client) {
	service := "0.0.0.0:21025"
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	rAddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listener.Accept()
		//fmt.Println("Server listerning")
		rConn, err := net.DialTCP("tcp", nil, rAddr)
		if err != nil {
			fmt.Println("\n[Error] Client tried to connect. Server is not accepting connections.")
			conn.Close()
		} else {
			fmt.Println("\nRemote addr:", rConn.LocalAddr())
			connections <- Client{"Unknown", -1, conn.RemoteAddr(), rConn.LocalAddr(), conn, rConn}
			go io.Copy(conn, rConn)
			go io.Copy(rConn, conn)
			defer rConn.Close()
		}
		defer conn.Close()
	}
}

type ServerInfo struct {
	Type, Data string
}

var prog string = "/home/rice/.starbound/linux64/starbound_server"

func monitorServer(cs chan ServerInfo) {
	for {
		cmd := exec.Command(prog)
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
func printMessages(count int) {
	if count == 0 {
		count = 20
	}
	path := "/home/rice/.starbound/linux64/starbound_server.log"
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Failed to read log file:", path)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if count > len(lines) {
		count = len(lines)
	}
	fmt.Println("Last", count, "log messages ( of", len(lines), "):")
	for i := 0; i < count; i++ {
		line := len(lines) - count + i
		fmt.Println(lines[line])
	}
	//return lines, scanner.Err()
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
		if command == "help" {
			printHelp()
		} else if command == "clients" {
			fmt.Println("Clients:")
			for i := 0; i < len(connections); i++ {
				conn := connections[i]
				fmt.Println(conn.Name, "- ID:", conn.Id, "IP:", conn.RemoteAddr)
			}
		} else if command == "log" {
			count := 20
			if len(parts) == 2 {
				count, _ = strconv.Atoi(parts[1])
			}
			printMessages(count)
		} else {
			fmt.Println("Unknown command:", command)
			printWTF()
		}
	}
}
func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("ban <name>     - Ban a player's IP by name.")
	fmt.Println("bans           - List all banned players and IPs.")
	fmt.Println("banip <ip>     - Ban a player by IP.")
	fmt.Println("broadcast      - Broadcast a message.")
	fmt.Println("clients        - Display connected clients.")
	fmt.Println("help           - This message.")
	fmt.Println("log [<count>]    - Display the last <count> server messages. <count> defaults to 20.")
	fmt.Println("unban <name>   - Unban a player's IP by name.")
	fmt.Println("unbanip <ip>   - Unban a player by IP.")
}
func printWTF() {
	fmt.Println("Type 'help' for more information.")
}

var connections []Client

func main() {
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
				fmt.Println("[ServerMon] Info:", info.Type, "Data:", info.Data)
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
							}
						}
					} else if op=="disconnected" {
						id_str := parts[len(parts)-3]
						id_str = id_str[1 : len(id_str)-1]
						id, _ := strconv.Atoi(id_str)
						for i := 0; i < len(connections); i++ {
							if connections[i].Id == id {
							    connections = append(connections[:i]...,connections[i+1:]...)
                                break
                            }
						}

                    }
				} else if info.Type == "serverup" {
					fmt.Println("Server listening for connections.")
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
