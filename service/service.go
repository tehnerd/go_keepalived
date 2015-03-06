package service

import (
	"fmt"
	"go_keepalived/healthchecks"
	"log/syslog"
	"strings"
)

type RealServer struct {
	RIP       string // real ip address
	Port      string
	Check     string // type of healthcheck
	Meta      string // meta info; for example: if need to be tunneled
	Proto     string // inhereted from service
	State     bool
	Timeout   int
	ToService chan ServiceMsg // inhereted from service; chan to communicate events to service container
	ToReal    chan ServiceMsg
}

type Service struct {
	VIP         string //Virtual IP address
	Proto       string
	Port        string
	State       bool
	AliveReals  int
	Quorum      int
	Hysteresis  int
	Timeout     int
	ToService   chan ServiceMsg
	FromService chan ServiceMsg
	FromReal    chan ServiceMsg
	logWriter   *syslog.Writer
	Reals       []RealServer
}

/*
	service container <-> realServer communication (for example "kill realServer")
	or realServer failed healthcheck etc
*/
type ServiceMsg struct {
	Cmnd string
	Data string
}

type ServicesList struct {
	List      []Service
	logWriter *syslog.Writer
}

/*
	ServicesList is a container for all services, which must be monitored by
	keepalived.
*/

func (sl *ServicesList) Init() {
	writer, err := syslog.New(syslog.LOG_INFO, "go_keepalived:")
	if err != nil {
		panic("cant connect to syslog")
	}
	sl.logWriter = writer
}

func (sl *ServicesList) Add(srvc Service) {
	for cntr := 0; cntr < len(sl.List); cntr++ {
		if sl.List[cntr].isEqual(&srvc) {
			logMsg := strings.Join([]string{"service already exists", srvc.VIP, " ",
				srvc.Proto, ":", srvc.Port}, " ")
			sl.logWriter.Write([]byte(logMsg))
			return
		}
	}
	srvc.logWriter = sl.logWriter
	sl.List = append(sl.List, srvc)
	logMsg := strings.Join([]string{"added new service", srvc.VIP, " ",
		srvc.Proto, ":", srvc.Port}, " ")
	sl.logWriter.Write([]byte(logMsg))

}

func (srvc *Service) isEqual(otherSrvc *Service) bool {
	if srvc.VIP == otherSrvc.VIP && srvc.Proto == otherSrvc.Proto &&
		srvc.Port == otherSrvc.Port {
		return true
	} else {
		return false
	}
}

//TODO: logic with  channel to kill service's goroutine
func (sl *ServicesList) Remove(srvc Service) {
	for cntr := 0; cntr < len(sl.List); cntr++ {
		if sl.List[cntr].isEqual(&srvc) {
			if cntr != len(sl.List)-1 {
				sl.List = append(sl.List[:cntr], sl.List[cntr+1:]...)
			} else {
				sl.List = sl.List[:cntr]
			}
			logMsg := strings.Join([]string{"removed service", srvc.VIP, " ",
				srvc.Proto, ":", srvc.Port}, " ")
			sl.logWriter.Write([]byte(logMsg))
		}
	}
}

/*
 Funcs for working with Service container and
 real servers inside Service container
*/

func (sl *ServicesList) Start() {
	for cntr := 0; cntr < len(sl.List); cntr++ {
		go sl.List[cntr].StartService()
	}
}

func (srvc *Service) Init() {
	srvc.Timeout = 1
	srvc.ToService = make(chan ServiceMsg)
	srvc.FromService = make(chan ServiceMsg)
	srvc.FromReal = make(chan ServiceMsg)
}

func (srvc *Service) StartService() {
	for cntr := 0; cntr < len(srvc.Reals); cntr++ {
		go srvc.Reals[cntr].StartReal()
	}
	loop := 1
	for loop == 1 {
		select {
		case msg := <-srvc.FromReal:
			switch msg.Cmnd {
			case "Dead":
				srvc.AliveReals--
				logMsg := strings.Join([]string{"real server", msg.Data, "now dead"}, " ")
				srvc.logWriter.Write([]byte(logMsg))
				if srvc.AliveReals < srvc.Quorum && srvc.State == true {
					srvc.State = false
					logMsg = strings.Join([]string{"turning down service", srvc.VIP,
						srvc.Port, srvc.Proto}, " ")
					srvc.logWriter.Write([]byte(logMsg))
				}
			case "Alive":
				srvc.AliveReals++
				logMsg := strings.Join([]string{"real server", msg.Data, "now alive"}, " ")
				srvc.logWriter.Write([]byte(logMsg))
				//TODO: hysteresis
				if srvc.AliveReals >= srvc.Quorum && srvc.State == false {
					srvc.State = true
					logMsg = strings.Join([]string{"bringing up service", srvc.VIP,
						srvc.Port, srvc.Proto}, " ")
					srvc.logWriter.Write([]byte(logMsg))
				}
			}
		}
	}
}

func (rlSrv *RealServer) StartReal() {
	fields := strings.Fields(rlSrv.Check)
	checkType := fields[0]
	/*
			we will send 1 to this chan when we are going to remove this real server
		 	and  turn off this check
	*/
	toCheck := make(chan int)
	/*
		we are going to receive feedback from the check subsystem thru this chan;
		for example 0 if check failed and 1 otherwise
	*/
	fromCheck := make(chan int)
	switch checkType {
	case "tcp":
		checkLine := []string{rlSrv.RIP, rlSrv.Port}
		go healthchecks.TCPCheck(toCheck, fromCheck, checkLine, rlSrv.Timeout)
	case "http", "https":
		if len(fields) < 2 {
			//TODO: logic to proper kill of the real servers goroutine
			return
		}
		checkLine := fields[1:]
		go healthchecks.HTTPCheck(toCheck, fromCheck, checkLine, rlSrv.Timeout)
	}
	loop := 1
	for loop == 1 {
		select {
		case result := <-fromCheck:
			switch result {
			case -1:
				//TODO: add logic to remove real coz of this error
				fmt.Println("cant resolve remote addr")
			case 1:
				if rlSrv.State == false {
					rlSrv.State = true
					DataString := ""
					if rlSrv.Meta != "" {
						//right now i can only think meta == tun; prob gonna rewrite this later
						DataString = strings.Join([]string{rlSrv.RIP, rlSrv.Meta}, " ")
					} else {
						DataString = strings.Join([]string{rlSrv.RIP, rlSrv.Port, rlSrv.Proto}, " ")
					}
					rlSrv.ToService <- ServiceMsg{Cmnd: "Alive", Data: DataString}
				}
			case 0:
				if rlSrv.State == true {
					rlSrv.State = false
					DataString := ""
					if rlSrv.Meta != "" {
						//right now i can only think meta == tun; prob gonna rewrite this later
						DataString = strings.Join([]string{rlSrv.RIP, rlSrv.Meta}, " ")
					} else {
						DataString = strings.Join([]string{rlSrv.RIP, rlSrv.Port, rlSrv.Proto}, " ")
					}
					rlSrv.ToService <- ServiceMsg{Cmnd: "Dead", Data: DataString}
				}
			}
		}
	}

}

func IsServiceValid(srvc Service) bool {
	if srvc.VIP != "" && srvc.Port != "" && srvc.Proto != "" {
		return true
	} else {
		return false
	}
}

func (rlsrv *RealServer) isEqualReal(otherRlsrv RealServer) bool {
	if rlsrv.RIP == otherRlsrv.RIP && rlsrv.Port == otherRlsrv.Port {
		return true
	} else {
		return false
	}
}

//TODO: add loging as well
func (srvc *Service) AddReal(rlsrv RealServer) {
	for cntr := 0; cntr < len(srvc.Reals); cntr++ {
		if srvc.Reals[cntr].isEqualReal(rlsrv) {
			return
		}
	}
	rlsrv.Proto = srvc.Proto
	rlsrv.Timeout = srvc.Timeout
	rlsrv.ToService = srvc.FromReal
	rlsrv.ToReal = make(chan ServiceMsg)
	srvc.Reals = append(srvc.Reals, rlsrv)

}

//TODO: add remove and change realServer logic
