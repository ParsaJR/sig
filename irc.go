package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"
)

const partmsg string = "Goodbye so soon"
var address string = "irc.libera.chat"
var port string = "6667"
var nick string = "Gonzoga"
var name string = os.Getenv("USER")
var ssl bool = false
var channel string = ""

func parseargs(args []string) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-a":
			if len(args) <= i+1 {
				fmt.Println("Invalid arguments!")
				os.Exit(2)
			}
			address = args[i+1]
			i++
			break
		case "-p":
			if len(args) <= i+1 {
				fmt.Println("Invalid arguments!")
				os.Exit(2)
			}
			port = args[i+1]
			i++
			break
		case "-ssl":
			ssl = true
			break
		case "-v":
			fmt.Println("sig: v1.0 Bryce Vandegrift")
			os.Exit(1)
			break
		case "-h":
			usage()
			os.Exit(1)
			break
		default:
			fmt.Println("Error: Invalid arguments!!!")
			os.Exit(2)
			break
		}
	}
}

func parsein(conn net.Conn, input string) {
	if input[0] == '\n' || input == "" {
		return
	}
	if input[0] != ':' {
		if channel == "" {
			fmt.Println("No channel to send to!")
		} else {
			msend(conn, "PRIVMSG " + channel + " :" + input)
			fmt.Fprintf(os.Stdout, "%s >< (%s): %s", channel, nick, input)
		}
		return
	}
	if input[1] != '\n' {
		arg := strings.Fields(input)
		switch input[1] {
		case 'j':
			if len(arg) < 2 {
				fmt.Println("Error: Not enough args!!!")
				break
			}
			channel = arg[1]
			msend(conn, "JOIN " + channel)
			fmt.Println("Joined " + channel)
			break
		case 'l':
			if len(arg) < 2 {
				fmt.Println("Error: Not enough args!!!")
				break
			}
			channel = ""
			msend(conn, "PART " + arg[1] + " " + partmsg)
			break
		case 'm':
			if len(arg) < 3 {
				fmt.Println("Error: Not enough args!!!")
				break
			}
			msend(conn, "PRIVMSG " + arg[1] + " :" + strings.Join(arg[2:], " "))
			break
		case 'h':
			fmt.Println("Commands: [j](join channel) [l](leave channel) [m](private message) [h](view this message) [q](quit)")
			break
		case 'q':
			os.Exit(0)
			break
		default:
			fmt.Println("Error: Invalid command!!!")
			break
		}
	}
}

func parseout(conn net.Conn, output string) {
	if output[1] == '\n' || output == "" {
		return
	}
	if output[0:4] == "PING" {
		msend(conn, "PONG :" + address)
		return
	}
	var data = output
	if output[0] == ':' {
		data = output[1:]
		delimString := strings.Fields(data)
		usr := strings.Split(delimString[0], "!")
		cmd := delimString[1]
		recp := delimString[2]
		text := strings.Join(delimString[3:], " ")

		fmt.Fprintf(os.Stdout, "%s >< %s %s %s\n", usr[0], cmd, recp, text)
		return
	}
	fmt.Println(data)
}

func connect(address string, port string, ssl bool) net.Conn {
	var conn net.Conn
	var err error
	if ssl {
		conn, err = tls.Dial("tcp", address + ":" + port, &tls.Config {})
	} else {
		conn, err = net.Dial("tcp", address + ":" + port)
	}
	if err != nil {
		fmt.Println("Could not connect to server!")
		os.Exit(2)
	}

	msend(conn, "NICK " + nick)
	msend(conn, "USER " + name + " 0 * :" + name)
	fmt.Println("Connected!")
	return conn
}

func msend(conn net.Conn, data string) {
	fmt.Fprintf(conn, data + "\r\n")
}

func uinput(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		parsein(conn, text)
	}
}

func usage() {
	fmt.Println("Usage: irc [-a host] [-p port] [-ssl] [-h] [-v]")
}

func main() {
	args := os.Args[1:]
	parseargs(args)

	connection := connect(address, port, ssl)
	defer connection.Close()
	status := bufio.NewScanner(connection)

	go uinput(connection)
	for status.Scan() {
		//fmt.Println(status.Text())
		parseout(connection, status.Text())
	}
}
