package main

// kompilacja packr build .
import (
	"bufio"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"
	//http json
	"encoding/json"
	//"log"
	"html/template"
	"net/http"

	"github.com/gobuffalo/packr"
	"github.com/julienschmidt/httprouter"
	"github.com/tatsushid/go-fastping"
)

var (
	to_print        chan string
	to_done         chan struct{}
	update_record   chan *Record
	new_record      chan *Record
	get_one_record  chan string
	send_one_record chan *Record
	send_records    chan map[string]*Record
	get_records     chan bool
	time_in_second  chan time.Duration
	box             packr.Box
	REST            *httprouter.Router = httprouter.New()
	//start_ping      chan bool
	//pozwol_ping     chan bool
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = "3333"
	CONN_TYPE = "tcp"
)

type HostHttp struct {
	Hostname string
	Rhost    string
}
type RecordJ struct {
	Nazwa_hosta      string `json:"nazwa_hosta"`
	Data_pomiaru     string `json:"data_pomiaru"`
	Status_pomiaru   int    `json:"status_pomiaru"`
	Status_wysylkowy int    `json:"status_wysyłkowy"`
	Czas_pingu       string `json:"czas_ping"`
	Licznik_testow   int    `json:"licznik_testow"`
	Licznik_bledow   int    `json:"licznik_bledow"`
}

type Record struct {
	Nazwa_hosta           string
	Data_ostatniego_testu time.Time
	Czas_unix             time.Time
	Czas_pingu            string
	Licznik_testow        int
	Licznik_bledow        int
	Ostatni_status        int
}

func NewRecord(nazwahosta string) *Record {
	return &Record{Nazwa_hosta: nazwahosta}
}

func (r *Record) UpdateRecord(nr *Record) {
	r.SetNazwaHosta(nr.Nazwa_hosta)
	r.SetData_ostatniego_testu(nr.Data_ostatniego_testu)
	r.SetCzas_unix(nr.Czas_unix)
	r.SetCzas_pingu(nr.Czas_pingu)
	r.SetLicznik_bledow(nr.Licznik_bledow)
	r.SetLicznik_testow(nr.Licznik_testow)
	r.SetOstatni_status(nr.Ostatni_status)
}
func (r Record) CopyToNewRecord() *Record {
	n := Record{}
	n = r
	return &n
}
func (r *Record) CopyToNewRecord_OLD() *Record {
	nr := &Record{}
	nr.SetNazwaHosta(r.Nazwa_hosta)
	nr.SetData_ostatniego_testu(r.Data_ostatniego_testu)
	nr.SetCzas_unix(r.Czas_unix)
	nr.SetCzas_pingu(r.Czas_pingu)
	nr.SetLicznik_bledow(r.Licznik_bledow)
	nr.SetLicznik_testow(r.Licznik_testow)
	nr.SetOstatni_status(r.Ostatni_status)
	return nr
}
func (r *Record) SetNazwaHosta(n string) {
	r.Nazwa_hosta = n
}
func (r *Record) SetData_ostatniego_testu(t time.Time) {
	r.Data_ostatniego_testu = t
}

func (r *Record) SetCzas_unix(t time.Time) {
	r.Czas_unix = t
}
func (r *Record) SetOstatni_status(i int) {
	r.Ostatni_status = i
}
func (r *Record) SetCzas_pingu(n string) {
	r.Czas_pingu = n
}
func (r *Record) SetLicznik_testow(i int) {
	r.Licznik_testow = i
}

func (r *Record) SetLicznik_bledow(i int) {
	r.Licznik_bledow = i
}
func (r *Record) AddLicznik_testow() {
	r.Licznik_testow++
}
func (r *Record) AddLicznik_bledow() {
	r.Licznik_bledow++
}

type ServiceConn struct {
	Con               net.Conn
	Number_connection int
	Ip_client         string
	To_local_done     chan struct{}
	To_stop_buffer    chan struct{}
	To_go             chan string
	To_go2            chan string
	To_send           chan string
	To_print_buffer   chan string
	To_work           chan string
	From_ping         chan int
	To_time_ping      chan string
}

func (s *ServiceConn) SetNumber_connection(numer int) {
	s.Number_connection = numer
}
func (s *ServiceConn) SetIp_client(ip string) {
	s.Ip_client = ip

}
func (s *ServiceConn) GoRunIsStopLocal() bool {
	select {
	case <-s.To_local_done:
		return true
	default:
		return false
	}
}
func (s *ServiceConn) Start() {
	s.To_go = make(chan string)
	s.To_go2 = make(chan string)
	s.To_send = make(chan string)
	s.To_local_done = make(chan struct{})
	s.To_print_buffer = make(chan string)
	s.To_stop_buffer = make(chan struct{})
	s.To_work = make(chan string)
	s.From_ping = make(chan int)
	s.To_time_ping = make(chan string)
	go s.PrintBuferr()
	go s.Sending_a_message()
	go s.Working()
	go s.Action()
	go s.Waiting_for_message()
}

func statistic_http(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	to_print <- "---- CLIENT " + r.RemoteAddr + " POBIERANIE STATYSTYK " + ps.ByName("host") + " ----\n\n"
	//fmt.Println(ps.ByName("host"))
	//HostHttp{hostname: ps.ByName("host"), rhost: r.Host}
	infotemplate := box.String("ping_statistic.html")
	t := template.New("statistic")
	t, _ = t.Parse(infotemplate)
	t.Execute(w, &HostHttp{Hostname: ps.ByName("host"), Rhost: r.Host})

}

func status_http(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	to_print <- "---- CLIENT " + r.RemoteAddr + " POBIERANIE DANYCH  ----\n\n"
	get_records <- true
	//data := make(map[string]*Record)
	data := <-send_records
	lista := []string{}
	lista2 := []RecordJ{}
	for k, _ := range data {
		lista = append(lista, k)
	}
	sort.Strings(lista)
	for i := 0; i < len(lista); i++ {
		Last_send := strconv.Itoa(data[lista[i]].Ostatni_status)
		if data[lista[i]].Ostatni_status == 1 {
			Last_send = strconv.FormatInt(time.Now().Unix()-data[lista[i]].Czas_unix.Unix(), 10)
		}
		recordj := RecordJ{}
		recordj.Nazwa_hosta = data[lista[i]].Nazwa_hosta
		recordj.Data_pomiaru = data[lista[i]].Data_ostatniego_testu.Format("2006-01-02 15:04:05")
		recordj.Status_pomiaru = data[lista[i]].Ostatni_status
		recordj.Status_wysylkowy, _ = strconv.Atoi(Last_send)
		recordj.Czas_pingu = data[lista[i]].Czas_pingu
		recordj.Licznik_testow = data[lista[i]].Licznik_testow
		recordj.Licznik_bledow = data[lista[i]].Licznik_bledow
		lista2 = append(lista2, recordj)
		//data[lista[i]].Nazwa_hosta,
		//						data[lista[i]].Data_ostatniego_testu.Format("2006-01-02 15:04:05"),
		//						strconv.Itoa(data[lista[i]].Ostatni_status),
		//						Last_send,
		//						data[lista[i]].Czas_pingu,
		//						strconv.Itoa(data[lista[i]].Licznik_testow),
		//						strconv.Itoa(data[lista[i]].Licznik_bledow))
		//type RecordJ struct {
		//	Nazwa_hosta           string
		//	Data_pomiaru          string
		//	Status_pomiaru        int
		//	Status_wysylkowy	  int
		//	Czas_pingu            string
		//	Licznik_testow        int
		//	Licznik_bledow        int
		//}

	}

	infoM, err := json.Marshal(lista2)
	if err != nil {
		fmt.Fprintf(w, "---- ERROR PRZY ZMIANIE STRUKTURY NA JSON ----")
	}
	w.Header().Set("Content-Type", "application/json ; charset=utf-8")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, string(infoM))

}

func ping_http(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	to_print <- "---- CLIENT " + r.RemoteAddr + " OTWIERANIE STRONY ----\n\n"
	infotemplate := box.String("ping.html")

	t := template.New("info")
	t, _ = t.Parse(infotemplate)
	t.Execute(w, r.Host)

}
func sluchaj_http() {
	//REST * httprouter.Router = httprouter.New()

	REST.GET("/pingStatus", status_http)
	REST.GET("/ping/:host", statistic_http)
	REST.GET("/ping", ping_http)
	//REST.GET("/test", ping_http)
	//	http.HandleFunc("/pingStatus", status_http) // setting router rule
	//	// setting router rule
	//	http.HandleFunc("/ping/", statistic_http)
	//	http.HandleFunc("/ping", ping_http)
	//	err := http.ListenAndServe(":9090", nil) // setting listening port
	//	//nie sprawdzam czy sie udała otworzyc port http dla stornki
	//	if err != nil {
	//		to_print <- "---- ERROR PRZY URUCHOMIENIU SERWERA HTTP----\n\n"
	//	}
	err := http.ListenAndServe(":80", REST)
	if err != nil {
		to_print <- "---- ERROR PRZY URUCHOMIENIU SERWERA HTTP----\n\n"
	}

}

func (s *ServiceConn) Waiting_for_message() {
	timeoutDuration := 5 * time.Second
	s.Con.SetReadDeadline(time.Now().Add(timeoutDuration))
	b := bufio.NewReader(s.Con)
	message, err := b.ReadString('\n')
	if err != nil {
		s.To_print_buffer <- "---- CLIENT NIE ZOSTAWIŁ WIADOMOSCI ----\n"
		s.To_print_buffer <- "Ip Clienta: " + s.Ip_client + "\n"
		s.To_print_buffer <- "Numer Połaczenia: " + strconv.Itoa(s.Number_connection) + "\n"
		s.To_print_buffer <- "INFO: Mogł wystąpić bład, lub upłynoł okreslony czas\n"
		s.To_print_buffer <- "\n"
		close(s.To_stop_buffer)
		//close(s.To_local_done)
		return
	}
	m := strings.TrimSpace(message)
	s.To_go <- m
	if m == "SETTIME" {
		message, err = b.ReadString('\n')
		if err != nil {
			s.To_print_buffer <- "---- CLIENT NIE ZOSTAWIŁ WIADOMOSCI ----\n"
			s.To_print_buffer <- "Ip Clienta: " + s.Ip_client + "\n"
			s.To_print_buffer <- "Numer Połaczenia: " + strconv.Itoa(s.Number_connection) + "\n"
			s.To_print_buffer <- "INFO: Mogł wystąpić bład, lub upłynoł okreslony czas\n"
			s.To_print_buffer <- "\n"
			close(s.To_stop_buffer)
			//close(s.To_local_done)
			return
		} else {
			m = strings.TrimSpace(message)
			s.To_go2 <- m
			return
		}

	} else {
		return
	}

}

func (s *ServiceConn) Action() {

	for !s.GoRunIsStopLocal() {
		select {
		case info := <-s.To_go:
			switch info {
			case "STOP":
				s.To_print_buffer <- "---- WYŁĄCZENIE CAŁEGO SERWERA----\n"
				s.To_print_buffer <- "Ip Clienta: " + s.Ip_client + "\n"
				s.To_print_buffer <- "Numer Połaczenia: " + strconv.Itoa(s.Number_connection) + "\n"
				s.To_send <- "SERWER SIĘ WYŁACZA\n"
				time.Sleep(5 * time.Millisecond)
				close(to_done)
				return
			case "SETTIME":
				s.To_print_buffer <- "---- ZMIANA OKRESU PRÓBKOWANIA ----\n"
				s.To_print_buffer <- "Ip Clienta: " + s.Ip_client + "\n"
				s.To_print_buffer <- "Numer Połaczenia: " + strconv.Itoa(s.Number_connection) + "\n"
				//go s.Waiting_for_message()
				select {
				case time_info := <-s.To_go2:
					t, er := strconv.ParseInt(time_info, 10, 64)
					if er != nil {
						s.To_print_buffer <- "NIE POPRAWNY PARAMETR CZASU\n"
						s.To_send <- "NIE POPRAWNY PARAMETR CZASU\n"
					} else {
						s.To_print_buffer <- "ZMIANA OKRESU NA: " + time_info + " Sekund\n"
						time_in_second <- time.Duration(t)
						s.To_send <- "OKRESU PRÓBKOWANIA ZMIENIONY\n"

					}
				case <-to_done:
					return
				case <-s.To_local_done:
					s.Con.Close()
					return
				}
				//s.To_send <- "SERWER PING RUNNING\n"
				//close(s.To_local_done)
			case "STATUS":
				s.To_print_buffer <- "---- POBRANIE STATUSU SERWERA ----\n"
				s.To_print_buffer <- "Ip Clienta: " + s.Ip_client + "\n"
				s.To_print_buffer <- "Numer Połaczenia: " + strconv.Itoa(s.Number_connection) + "\n"
				s.To_send <- "SERWER PING RUNNING\n"
				//close(s.To_local_done)
			case "INFO":
				s.To_print_buffer <- "---- POBRANIE STATUSU WSZYSTKICH KONT ----\n"
				s.To_print_buffer <- "Ip Clienta: " + s.Ip_client + "\n"
				s.To_print_buffer <- "Numer Połaczenia: " + strconv.Itoa(s.Number_connection) + "\n"
				get_records <- true
				//data := make(map[string]*Record)
				data := <-send_records
				lista := []string{}
				for k, _ := range data {
					lista = append(lista, k)
				}
				sort.Strings(lista)
				infoAll := fmt.Sprintf("%-30s%-30s%-20s%-20s%-20s%-20s%-20s\n", "Nazwa Hosta", "Data Pomiaru", "Status Pomiaru", "Status Wysyłkowy", "Czas Pingu", "Licznik Testow", "Licznik Błędów")
				for i := 0; i < len(lista); i++ {
					Last_send := strconv.Itoa(data[lista[i]].Ostatni_status)
					if data[lista[i]].Ostatni_status == 1 {
						Last_send = strconv.FormatInt(time.Now().Unix()-data[lista[i]].Czas_unix.Unix(), 10)
					}
					infoAdd := fmt.Sprintf("%-30s%-30s%-20s%-20s%-20s%-20s%-20s\n",
						data[lista[i]].Nazwa_hosta,
						data[lista[i]].Data_ostatniego_testu.Format("2006-01-02 15:04:05"),
						strconv.Itoa(data[lista[i]].Ostatni_status),
						Last_send,
						data[lista[i]].Czas_pingu,
						strconv.Itoa(data[lista[i]].Licznik_testow),
						strconv.Itoa(data[lista[i]].Licznik_bledow))
					infoAll = infoAll + infoAdd
				}
				infoAll = infoAll + "\r"
				s.To_send <- infoAll
				//fmt.Sprintf("%-20s%-20s%-20s%-20s%-20s%-20s\n",

				//s.To_send <- "SERWER PING RUNNING\n"
				//close(s.To_local_done)
			default:
				s.To_print_buffer <- "---- TEST PING: " + info + " ----\n"
				s.To_print_buffer <- "Ip Clienta: " + s.Ip_client + "\n"
				s.To_print_buffer <- "Numer Połaczenia: " + strconv.Itoa(s.Number_connection) + "\n"
				s.To_work <- info
			}
		case <-to_done:
			return
		case <-s.To_local_done:
			s.Con.Close()
			return
		}
	}
	s.Con.Close()
}

func (s *ServiceConn) PrintBuferr() {
	m := ""
	for !s.GoRunIsStopLocal() {
		select {
		case info := <-s.To_print_buffer:
			m = m + info
		case <-s.To_stop_buffer:
			to_print <- m
			close(s.To_local_done)
			return
		case <-to_done:
			return
		}
	}

}
func (s *ServiceConn) Working() {
	select {
	case hostname := <-s.To_work:

		go s.PingTest(hostname)
		ping_status := <-s.From_ping
		ping_rtt := <-s.To_time_ping
		ping_status_str := strconv.Itoa(ping_status)
		s.To_print_buffer <- "Data testu: " + time.Now().Format("2006-01-02 15:04:05") + "\n"
		if ping_status == -2 {
			s.To_print_buffer <- "INFO: BŁAD - nie uruchomiono serwera jako root\n"

			//s.To_send <- ping_status_str + "\n"
			//return
		}
		if ping_status == -1 {
			s.To_print_buffer <- "INFO: Nazwa Hosta Nie prawidłowa\n"
			//Wtedy gdy nie rozpznaje hosta
			//s.To_print_buffer <- "\n"
			//s.To_send <- ping_status_str + "\n"

			//return
		}

		if ping_status == 0 {
			s.To_print_buffer <- "INFO: Nazwa Hosta Prawidłowa\n"
			s.To_print_buffer <- "INFO: Host pinguje\n"
			s.To_print_buffer <- "Czas Odpowiedzi Ping: " + ping_rtt + "\n"
			s.To_print_buffer <- "Status pomiaru: " + ping_status_str + "\n"
			get_one_record <- hostname
			record_in_basa := <-send_one_record
			record_in_basa.SetData_ostatniego_testu(time.Now())
			record_in_basa.SetCzas_unix(time.Now())
			record_in_basa.SetCzas_pingu(ping_rtt)
			record_in_basa.AddLicznik_testow()
			record_in_basa.SetOstatni_status(0)
			if record_in_basa.Licznik_testow == 1 {

				new_record <- record_in_basa
			} else {
				update_record <- record_in_basa //.CopyToNewRecord()

			}
			//s.To_print_buffer <- "\n"
			//s.To_send <- ping_status_str + "\n"
			//return

		}

		if ping_status == 1 {
			s.To_print_buffer <- "INFO: Nazwa Hosta Prawidłowa\n"
			s.To_print_buffer <- "INFO: Host nie pinguje\n"
			s.To_print_buffer <- "Status pomiaru: " + ping_status_str + "\n"
			get_one_record <- hostname
			record_in_basa := <-send_one_record
			record_in_basa.SetData_ostatniego_testu(time.Now())
			record_in_basa.SetCzas_pingu(ping_rtt)
			record_in_basa.AddLicznik_testow()
			record_in_basa.AddLicznik_bledow()

			if record_in_basa.Licznik_testow == 1 {
				record_in_basa.SetCzas_unix(time.Now())
				record_in_basa.SetOstatni_status(1)
				new_record <- record_in_basa
			} else {
				if record_in_basa.Ostatni_status == 0 {
					record_in_basa.SetOstatni_status(1)

				} else if record_in_basa.Ostatni_status == 1 {
					ping_status_str = strconv.FormatInt(time.Now().Unix()-record_in_basa.Czas_unix.Unix(), 10)
				}
				update_record <- record_in_basa

			}
			//s.To_print_buffer <- "\n"
			//s.To_send <- ping_status_str + "\n"
			//return
		}

		s.To_print_buffer <- "Status wysłany: " + ping_status_str + "\n"
		s.To_print_buffer <- "\n"
		s.To_send <- ping_status_str + "\n"
		return

	case <-to_done:
		return
	case <-s.To_local_done:

		return
	}

}

func (s *ServiceConn) PingTest3(hostname string, typ_p string) (status_tcp int, stan_rtt_tcp string) {
	t := fastping.NewPinger()

	status_tcp = -1

	stan_rtt_tcp = "error"
	rat, err := net.ResolveIPAddr("ip4:"+typ_p, hostname)
	if err != nil {
		status_tcp = -1
		stan_rtt_tcp = "error"

	} else {
		t.AddIPAddr(rat)
		status_tcp = 1
		stan_rtt_tcp = "nie_pinguje"
		t.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
			status_tcp = 0
			stan_rtt_tcp = rtt.String()
		}
		err = t.Run()
		if err != nil {
			s.To_print_buffer <- err.Error() + "\n"
			status_tcp = -2
			stan_rtt_tcp = "error"
		}

	}
	return status_tcp, stan_rtt_tcp
}
func (s *ServiceConn) PingTest(hostname string) {
	status := -100
	rtt := ""
loop:
	for !s.GoRunIsStopLocal() {

		//		status_tcp, _ := s.PingTest3(hostname, "tcp")
		status_udp, _ := s.PingTest3(hostname, "udp")
		status_icmp, rtt_icmp := s.PingTest3(hostname, "icmp")
		//fmt.Println(status_tcp, status_udp, status_icmp)
		if status_icmp == status_udp {
			//		if (status_icmp == status_tcp) == (status_tcp == status_udp) {
			status = status_icmp
			rtt = rtt_icmp
			break loop
		}
	}
	s.From_ping <- status
	s.To_time_ping <- rtt

}

func (s *ServiceConn) Sending_a_message() {
	for !s.GoRunIsStopLocal() {
		select {
		case info := <-s.To_send:

			_, err := s.Con.Write([]byte(info))
			if err != nil {
				s.Con.Close()
				s.To_print_buffer <- "Write to server failed: " + err.Error()
			}
			close(s.To_stop_buffer)
			return
		case <-to_done:
			return
		case <-s.To_local_done:

			return
		case <-time.After(20 * time.Second):
			s.To_print_buffer <- "ERROR: KLIENT NIE OTRZYMAŁ ODPOWIEDZI MINOŁ CZAS 20s\n"
			//start_ping <- false
			close(s.To_stop_buffer)
			return

		}
	}
}

func main() {
	MakeGlobalChan()
	//go PulaPing()
	box = packr.NewBox("./templates")
	go Printer()
	go AccesToData()
	go sprawdzaj_hosty()

	StartSerwer()

}

func StartSerwer() {
	server, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if server == nil {
		//to_print <- "SERWER NIE WYSTARTOWAŁ BŁAD:/n" + err.Error()
		panic("SERWER MA ZAJETY PORT: " + err.Error())
	}
	go sluchaj_http()
	//fmt.Println("test")
	conns := clientConns(server)

	for true {
		select {
		case nowe_polaczenie := <-conns:
			go nowe_polaczenie.Start()
		case <-to_done:
			return
		}
	}

}

func clientConns(listener net.Listener) chan ServiceConn {
	ch := make(chan ServiceConn, 255)
	i := 0
	go func() {
		for !GoRunIsStop() {

			client, err := listener.Accept()
			NewServiceCon := ServiceConn{Con: client}

			if NewServiceCon.Con == nil {
				//to_print <- "CLIENT NIE MOZE SIE POŁACZYĆ:\n" + err.Error()
				panic("couldn't accept: " + err.Error())
				continue
			}
			i++
			NewServiceCon.SetNumber_connection(i)
			NewServiceCon.SetIp_client(client.RemoteAddr().String())
			//fmt.Printf("%d: %v <-> %v\n", i, client.LocalAddr(), client.RemoteAddr())

			ch <- NewServiceCon
		}
	}()
	return ch
}

func MakeGlobalChan() {
	to_print = make(chan string, 255)
	to_done = make(chan struct{})
	//start_ping = make(chan bool)
	//pozwol_ping = make(chan bool)
	update_record = make(chan *Record, 255)
	new_record = make(chan *Record, 255)
	send_one_record = make(chan *Record)
	get_one_record = make(chan string)
	send_records = make(chan map[string]*Record, 255)
	get_records = make(chan bool)
	time_in_second = make(chan time.Duration)

}
func Printer() {
	for {
		select {
		case info := <-to_print:
			fmt.Println(info)
		case <-to_done:
			return
		}
	}
}
func sprawdzaj_hosty() {
	var okres_time time.Duration
	okres_time = 2
loop:
	for true {
		select {
		case <-to_done:
			break loop
		case t := <-time_in_second:
			okres_time = t

		case <-time.After(okres_time * time.Second):
			get_records <- true
			new_mapa_danych := <-send_records
			for k, _ := range new_mapa_danych {
				time.Sleep(30 * time.Millisecond)
				go sprawdz_hosta(k)
				//new_mapa_danych[k] = v
			}
		}
	}
}
func sprawdz_hosta(hostname string) int {

	conn, err := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	//conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		// właczmy serwer
		return 1
	} else {
		// zapytajmy serwer czy to nasz
		b := bufio.NewReader(conn)
		timeoutDuration := 20 * time.Second
		conn.Write([]byte(hostname + "\n"))
		conn.SetReadDeadline(time.Now().Add(timeoutDuration))
		_, err := b.ReadString('\n')
		conn.Close()
		if err != nil {
			return 1
		} else {
			return 0
		}
	}
}
func AccesToData() {
	Dane := make(map[string]*Record)
	for {
		select {
		case upd_r := <-update_record:
			Dane[upd_r.Nazwa_hosta].UpdateRecord(upd_r)
		case new_r := <-new_record:
			Dane[new_r.Nazwa_hosta] = new_r //.CopyToNewRecord()
		case record_name := <-get_one_record:
			_, ok := Dane[record_name]
			if ok {
				send_one_record <- Dane[record_name] //.CopyToNewRecord()
			} else {
				send_one_record <- NewRecord(record_name)

			}
			//to_print <- Dane[record_name].Nazwa_hosta
			//to_print <- strconv.Itoa(Dane[record_name].Licznik_testow)
		case <-get_records:
			Nowa := make(map[string]*Record)
			for k, v := range Dane {
				Nowa[k] = v.CopyToNewRecord()
			}
			send_records <- Nowa
			//			Nowa := make(map[string]*Record)
			//			lista := []string{}
			//			for k, _ := range Dane {
			//				lista = append(lista, k)
			//			}
			//			sort.Strings(lista)
			//			for i := 0; i < len(lista); i++ {
			//				Nowa[lista[i]] = Dane[lista[i]].CopyToNewRecord()
			//			}
			//			send_records <- Nowa

		case <-to_done:
			return
		}
	}

}
func GoRunIsStop() bool {
	select {
	case <-to_done:
		return true
	default:
		return false
	}
}
