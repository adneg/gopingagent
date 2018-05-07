package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = "3333"
	CONN_TYPE = "tcp"
)

func main() {
	switch ile_argumentow := len(os.Args); ile_argumentow {
	case 1:
		Run1()
	case 2:
		Run2()
	case 3:
		Run3()
	default:
	}
}

func Run3() {
	switch sluchanie := sprawdz_port(); sluchanie {
	case -1:
		fmt.Println("NIE MOŻNA NIC ZROBIĆ\nPORT ZAJĘTY")
	case 0:

		//fmt.Println("URUCHAMIANIE SERWERA")
		demonize(os.Args[0] + "S")
		time.Sleep(1 * time.Second)
		if os.Args[1] == "SETTIME" {

		} else {
			fmt.Println("ŹLE DOBRANE PARAMETRY")

		}

	case 1:
		//fmt.Println("SERWER URUCHOMIONY")
		if os.Args[1] == "SETTIME" {
			fmt.Println(set_time())
		} else {
			fmt.Println("ŹLE DOBRANE PARAMETRY")

		}
	}

}

func Run2() {
	switch sluchanie := sprawdz_port(); sluchanie {
	case -1:
		fmt.Println("NIE MOŻNA NIC ZROBIĆ\nPORT ZAJĘTY")
	case 0:

		//fmt.Println("URUCHAMIANIE SERWERA")
		demonize(os.Args[0] + "S")
		time.Sleep(1 * time.Second)
		if os.Args[1] == "INFO" {
			fmt.Println(sprawdz_info(os.Args[1]))
		} else {
			fmt.Println(sprawdz_hosta(os.Args[1]))
		}
	case 1:
		//fmt.Println("SERWER URUCHOMIONY")
		if os.Args[1] == "INFO" {
			fmt.Println(sprawdz_info(os.Args[1]))
		} else {
			fmt.Println(sprawdz_hosta(os.Args[1]))
		}
	}

}

func Run1() {
	switch sluchanie := sprawdz_port(); sluchanie {
	case -1:
		fmt.Println("NIE MOŻNA NIC ZROBIĆ\nPORT ZAJĘTY")
	case 0:

		fmt.Println("URUCHAMIANIE SERWERA..")
		demonize(os.Args[0] + "S")

	case 1:
		fmt.Println("SERWER URUCHOMIONY..")
	}

}

func RunS() {
	switch sluchanie := sprawdz_port(); sluchanie {
	case -1:
		fmt.Println("NIE MOŻNA NIC ZROBIĆ")
	case 0:

		fmt.Println("URUCHAMIANIE SERWERA")
		demonize(os.Args[0] + "S")

	case 1:
		fmt.Println("SERWER URUCHOMIONY")
	}

}

func sprawdz_port() int {

	conn, err := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	//conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		// właczmy serwer
		return 0
	} else {
		// zapytajmy serwer czy to nasz
		b := bufio.NewReader(conn)
		timeoutDuration := 2 * time.Second
		conn.Write([]byte("STATUS\n"))
		conn.SetReadDeadline(time.Now().Add(timeoutDuration))
		message, err := b.ReadString('\n')

		conn.Close()
		if err != nil {
			//ten serwer nie odpowiada!
			return -1
		} else {
			switch odp := strings.TrimSpace(message); odp {
			case "SERWER PING RUNNING":
				//to nasz serwer
				return 1
			default:
				//ten serwer zle odpowiada!
				return -1
			}

		}
	}
}
func sprawdz_hosta(hostname string) string {

	conn, err := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	//conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		// właczmy serwer
		return "SERWER WYŁACZONY"
	} else {
		// zapytajmy serwer czy to nasz
		b := bufio.NewReader(conn)
		timeoutDuration := 20 * time.Second
		conn.Write([]byte(hostname + "\n"))
		conn.SetReadDeadline(time.Now().Add(timeoutDuration))
		message, err := b.ReadString('\n')
		conn.Close()
		if err != nil {
			return "BŁAD POŁACZENIA"
		} else {
			return strings.TrimSpace(message)
		}
	}
}
func set_time() string {

	conn, err := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	//conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		// właczmy serwer
		return "SERWER WYŁACZONY"
	} else {
		// zapytajmy serwer czy to nasz
		b := bufio.NewReader(conn)
		timeoutDuration := 20 * time.Second
		//fmt.Println(os.Args[2])
		conn.Write([]byte(os.Args[1] + "\n"))
		conn.Write([]byte(os.Args[2] + "\n"))
		conn.SetReadDeadline(time.Now().Add(timeoutDuration))
		message, err := b.ReadString('\n')
		conn.Close()
		if err != nil {
			return "BŁAD POŁACZENIA"
		} else {
			return strings.TrimSpace(message)
		}
	}
}
func sprawdz_info(hostname string) string {

	conn, err := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	//conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		// właczmy serwer
		return "SERWER WYŁACZONY"
	} else {
		// zapytajmy serwer czy to nasz
		b := bufio.NewReader(conn)
		timeoutDuration := 20 * time.Second
		conn.Write([]byte(hostname + "\n"))
		conn.SetReadDeadline(time.Now().Add(timeoutDuration))
		message, err := b.ReadString('\r')
		conn.Close()
		if err != nil {
			return "BŁAD POŁACZENIA"
		} else {
			return strings.TrimSpace(message)
		}
	}
}
func demonize(komenda string) error {
	outfile, err := os.Create("/var/log/ping_agent_log")
	cmd := exec.Command(komenda)

	//cmd := exec.Command("/home/kamil/work/src/projekty/daemon_ping/daemon_ping", "START")
	cmd.Stdin = os.Stdin
	cmd.Stdout = outfile
	cmd.Stderr = outfile
	err = cmd.Start()
	return err

}
