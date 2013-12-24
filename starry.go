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
    "encoding/json"
)

var serverAddress string = "localhost:21024"
var proxyAddress string = "0.0.0.0:21025"
type Client struct {
	Name                  string
	Id                    int
	RemoteAddr, LocalAddr net.Addr
	Conn, ProxyConn       net.Conn
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
        for i:=0;i<len(bans);i++ {
            ban := bans[i]
            if(strings.Index(conn.RemoteAddr().String(), ban.Addr)!=-1){
                fmt.Println("[Info] Banned client tried to connect from IP:",conn.RemoteAddr().String(),"Matched Rule Name:",ban.Name,"Rule IP:",ban.Addr)
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
                go io.Copy(conn, rConn)
                go io.Copy(rConn, conn)
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
		} else if command == "bans" {
			fmt.Println("[Bans]")
			for i := 0; i < len(bans); i++ {
				conn := bans[i]
				fmt.Println("  Name:", conn.Name, "IP:", conn.Addr)
			}
        } else if command == "banip" {
            if len(parts)>1 {
                desc := "None"
                if len(parts)>2 {
                    desc = strings.Join(parts[2:]," ")
                }
                fmt.Println("[Banned] ip:",parts[1], "Desc:", desc)
                bans = append(bans, Ban{parts[1],desc})
                for i:=0; i< len(connections);i++ {
                    conn := connections[i]
                    addr_bits := strings.Split(conn.RemoteAddr.String(),":")
                    addr := strings.Join(addr_bits[:len(addr_bits)-1],":")
                    if strings.Index(addr, parts[1])!=-1 {
                        fmt.Println("[Kicked] Name:", conn.Name, "IP:", addr)
                        conn.Conn.Close()
                        conn.ProxyConn.Close()
                    }
                }
                writeConfig()
            } else {
                fmt.Println("Invalid syntax.")
                printWTF()
            }
        } else if command == "unbanip" {
            if len(parts)>1 {
                //bans = append(bans, Ban{parts[1],"None"})
                for i:=0; i< len(bans);i++ {
                    if bans[i].Addr == parts[1]{
                        conn := bans[i]
                        fmt.Println("[Unbanned] Name:", conn.Name, "IP:", conn.Addr)
                        bans = append(bans[:i],bans[i+1:]...)
                        writeConfig()
                    } else if strings.Index(bans[i].Addr, parts[1])!=-1 {
                        fmt.Println("Did you mean:",bans[i].Name,"instead? ( IP:", bans[i].Addr, ")")
                    }
                }
            } else {
                fmt.Println("Invalid syntax.")
                printWTF()
            }
        } else if command == "ban" {
            if len(parts)>1 {
                //bans = append(bans, Ban{parts[1],"None"})
                name := strings.Join(parts[1:]," ")
                for i:=0; i< len(connections);i++ {
                    conn := connections[i]
                    addr_bits := strings.Split(conn.RemoteAddr.String(),":")
                    addr := strings.Join(addr_bits[:len(addr_bits)-1],":")
                    if conn.Name == name {
                        fmt.Println("[Banned] Name:", conn.Name, "IP:", addr)
                        bans = append(bans, Ban{addr, conn.Name})
                        conn.Conn.Close()
                        conn.ProxyConn.Close()
                        writeConfig()
                        //bans = append(bans[:i],bans[i+1:]...)
                    } else if strings.Index(conn.Name, name)!=-1 {
                        fmt.Println("Did you mean:",conn.Name,"( IP:", addr, ")")
                    }
                }
            } else {
                fmt.Println("Invalid syntax.")
                printWTF()
            }
        } else if command == "unban" {
            if len(parts)>1 {
                //bans = append(bans, Ban{parts[1],"None"})
                name := strings.Join(parts[1:]," ")
                for i:=0; i< len(bans);i++ {
                    if strings.Index(bans[i].Name, name)!=-1 {
                        conn := bans[i]
                        fmt.Println("[Unbanned] Name:", conn.Name, "IP:", conn.Addr)
                        bans = append(bans[:i],bans[i+1:]...)
                        writeConfig()
                        break
                    }
                }
            } else {
                fmt.Println("Invalid syntax.")
                printWTF()
            }
        } else if command == "clients" {
			fmt.Println("[Clients]")
			for i := 0; i < len(connections); i++ {
				conn := connections[i]
				fmt.Println(" ",conn.Name, "- ID:", conn.Id, "IP:", conn.RemoteAddr)
			}
		} else if command == "log" {
			count := 20
			if len(parts) == 2 {
				count, _ = strconv.Atoi(parts[1])
			}
			printMessages(count)
		} else {
            if len(command) > 0{
                fmt.Println("Unknown command:", command)
            }
			printWTF()
		}
	}
}
func printHelp() {
	fmt.Println("[Commands]")
    fmt.Println("  General:")
	fmt.Println("    clients\n      - Display connected clients.")
	fmt.Println("    broadcast\n      - Send a message to all connected players. WIP")
	fmt.Println("    help\n      - This message.")
	fmt.Println("    log [<count>]\n      - Display the last <count> server messages. <count> defaults to 20.")
	fmt.Println("  Banning:")
    fmt.Println("    bans\n      - List all banned players and IPs.")
	fmt.Println("    ban <name>\n      - Ban a currently connected player's IP by name.")
	fmt.Println("    banip <ip> [<name/desc>]\n      - Ban a player by IP. You can ban subnets by omitting the end of a address. Ex: 'ban 8.8.8.'")
	fmt.Println("    unban <name/desc>\n      - Unban a IP by name or description.")
	fmt.Println("    unbanip <ip>\n      - Unban a player by IP.")
}
func printWTF() {
	fmt.Println("Type 'help' for more information.")
}

type Ban struct {
    Addr, Name string
}

var bans        []Ban
var connections []Client

type Config struct {
    ServerPath string
    ServerAddress string
    ProxyAddress string
    Bans []Ban
}

func writeConfig() {
    config := Config{serverPath, serverAddress, proxyAddress, bans}
    b, err := json.MarshalIndent(config, "", "    ")
    if err != nil {
        fmt.Println("[Error] Failed to create JSON config.")
    }
    writeLines([]string{string(b)},"starry.config")
}
func readConfig() {
    lines, err := readLines("starry.config")
    if err!=nil {
        fmt.Println("[Error]",err)
        writeConfig()
    } else {
        var config Config
        err := json.Unmarshal([]byte(strings.Join(lines,"\n")),&config)
        if err!=nil {
            fmt.Println("[Error]",err)
        } else {
            serverPath = config.ServerPath
            serverAddress = config.ServerAddress
            proxyAddress = config.ProxyAddress
            bans = config.Bans
        }
    }
}

func readLines(path string) ([]string, error) {
  file, err := os.Open(path)
  if err != nil {
    return nil, err
  }
  defer file.Close()

  var lines []string
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    lines = append(lines, scanner.Text())
  }
  return lines, scanner.Err()
}

// writeLines writes the lines to the given file.
func writeLines(lines []string, path string) error {
  file, err := os.Create(path)
  if err != nil {
    return err
  }
  defer file.Close()

  w := bufio.NewWriter(file)
  for _, line := range lines {
    fmt.Fprintln(w, line)
  }
  return w.Flush()
}

func main() {
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
							    connections = append(connections[:i],connections[i+1:]...)
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
