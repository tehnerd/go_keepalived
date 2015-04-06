package service

import (
	"fmt"
	"go_keepalived/adapter"
	"go_keepalived/healthchecks"
	"go_keepalived/notifier"
	"log/syslog"
	"strconv"
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
	Weight    string
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
	ToAdapter   chan adapter.AdapterMsg
	logWriter   *syslog.Writer
	Reals       []RealServer
}

/*
	service container <-> realServer communication (for example "kill realServer")
	or realServer failed healthcheck etc
	Type of cmnds:
		"Alive" (RS->Service) if healtcheck was successful and realServer is working
			Data in this case would be RIP+Port+Proto in case of lack of metainfo
			and RIP+Meta if meta exists
		"Dead" (RS->Service) if healthcheck was failed and realServer is not working
			Data will be the same as above
		"RSFatalError" (RS->Service) if there was a fatal error during setup of real server
			(for example we wasnt able to resolve real's ip+port etc)
		"Shutdown" (Service->RS) if we want to remove realServer from service context
*/
type ServiceMsg struct {
	Cmnd    string
	Data    string
	DataMap *map[string]string
}

type ServicesList struct {
	List            []Service
	ToAdapter       chan adapter.AdapterMsg
	FromAdapter     chan adapter.AdapterMsg
	ToNotifier      chan notifier.NotifierMsg
	FromNotifier    chan notifier.NotifierMsg
	ToServiceList   chan ServiceMsg
	FromServiceList chan ServiceMsg
	Testing         bool
	logWriter       *syslog.Writer
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
	sl.ToAdapter = make(chan adapter.AdapterMsg)
	sl.FromAdapter = make(chan adapter.AdapterMsg)
	sl.ToNotifier = make(chan notifier.NotifierMsg)
	sl.FromNotifier = make(chan notifier.NotifierMsg)
	sl.ToServiceList = make(chan ServiceMsg)
	sl.FromServiceList = make(chan ServiceMsg)
	go adapter.StartAdapter(sl.ToAdapter, sl.FromAdapter, sl.ToNotifier)
	sl.logWriter = writer
}

func (sl *ServicesList) Add(srvc Service) error {
	for cntr := 0; cntr < len(sl.List); cntr++ {
		if sl.List[cntr].isEqual(&srvc) {
			logMsg := strings.Join([]string{"service already exists", srvc.VIP, " ",
				srvc.Proto, ":", srvc.Port}, " ")
			sl.logWriter.Write([]byte(logMsg))
			return fmt.Errorf("service already exists")
		}
	}
	srvc.logWriter = sl.logWriter
	srvc.ToAdapter = sl.ToAdapter
	sl.List = append(sl.List, srvc)
	logMsg := strings.Join([]string{"added new service", srvc.VIP, " ",
		srvc.Proto, ":", srvc.Port}, " ")
	sl.logWriter.Write([]byte(logMsg))
	return nil
}

func (sl *ServicesList) AddNotifier(notifierCfg notifier.NotifierConfig) {
	go notifier.Notifier(sl.ToNotifier, sl.FromNotifier, notifierCfg)
	for _, neighbour := range notifierCfg.NeighboursList {
		sl.ToNotifier <- notifier.NotifierMsg{Type: "AddPeer", Data: neighbour}
	}
}

func (srvc *Service) isEqual(otherSrvc *Service) bool {
	if srvc.VIP == otherSrvc.VIP && srvc.Proto == otherSrvc.Proto &&
		srvc.Port == otherSrvc.Port {
		return true
	} else {
		return false
	}
}

func (sl *ServicesList) FindService(srvc *Service) int {
	for num, localSrvc := range sl.List {
		if localSrvc.isEqual(srvc) {
			return num
		}
	}
	return -1
}

func (sl *ServicesList) Remove(srvc Service) error {
	for cntr := 0; cntr < len(sl.List); cntr++ {
		if sl.List[cntr].isEqual(&srvc) {
			removedSrvc := sl.List[cntr]
			if cntr != len(sl.List)-1 {
				sl.List = append(sl.List[:cntr], sl.List[cntr+1:]...)
			} else {
				sl.List = sl.List[:cntr]
			}
			removedSrvc.ToService <- ServiceMsg{Cmnd: "RemoveService"}
			logMsg := strings.Join([]string{"removed service", srvc.VIP, " ",
				srvc.Proto, ":", srvc.Port, " from services list"}, " ")
			sl.logWriter.Write([]byte(logMsg))
			return nil
		}
	}
	return fmt.Errorf("service doesnt exists\n")
}

func (sl *ServicesList) Start() {
	if sl.Testing {
		sl.ToAdapter <- adapter.AdapterMsg{Type: "StartTesting"}
	}
	for cntr := 0; cntr < len(sl.List); cntr++ {
		go sl.List[cntr].StartService()
	}
	go DispatchSLMsgs(sl)
}

/*
 Funcs for working with Service container and
 real servers inside Service container
 TODO: move to sep file?
*/

func (srvc *Service) Init() {
	srvc.Timeout = 1
	srvc.ToService = make(chan ServiceMsg)
	srvc.FromService = make(chan ServiceMsg)
	srvc.FromReal = make(chan ServiceMsg)
}

func (srvc *Service) FindRealFromMsgData(msgData string) *RealServer {
	//right now we always have msg data in format "rip port (meta)?"
	dataFields := strings.Fields(msgData)
	if len(dataFields) != 3 {
		dataFields = append(dataFields, "")
	}
	rlSrv, _ := srvc.FindReal(dataFields[0], dataFields[1], dataFields[2])
	return rlSrv
}

func (srvc *Service) FindReal(RIP, Port, Meta string) (*RealServer, int) {
	for cntr := 0; cntr < len(srvc.Reals); cntr++ {
		rlSrv := &(srvc.Reals[cntr])
		if rlSrv.RIP == RIP && rlSrv.Port == Port && rlSrv.Meta == Meta {
			return rlSrv, cntr
		}
	}
	return nil, 0
}

func (srvc *Service) RemoveReal(rlSrv *RealServer, index int, notifyReal bool) {
	if notifyReal {
		rlSrv.ToReal <- ServiceMsg{Cmnd: "Shutdown"}
	}
	//TODO: add logic w/ adapter (adapter is responsible for adding and removing datapath to real)
	if index == len(srvc.Reals) {
		srvc.Reals = srvc.Reals[:index]
	} else {
		srvc.Reals = append(srvc.Reals[:index], srvc.Reals[index+1:]...)
	}
	logMsg := strings.Join([]string{"removing real server", rlSrv.RIP, rlSrv.Port,
		rlSrv.Meta, "for service", srvc.VIP, srvc.Port, srvc.Proto}, " ")
	srvc.logWriter.Write([]byte(logMsg))
}

func GenerateAdapterMsg(msgType string, srvc *Service, rlSrv *RealServer) adapter.AdapterMsg {
	adapterMsg := adapter.AdapterMsg{}
	adapterMsg.Type = msgType
	adapterMsg.ServiceVIP = srvc.VIP
	adapterMsg.ServicePort = srvc.Port
	adapterMsg.ServiceProto = srvc.Proto
	if rlSrv != nil {
		adapterMsg.RealServerRIP = rlSrv.RIP
		adapterMsg.RealServerPort = rlSrv.Port
		adapterMsg.RealServerWeight = rlSrv.Weight
		adapterMsg.RealServerMeta = rlSrv.Meta
	}
	return adapterMsg
}

func (srvc *Service) StartService() {
	for cntr := 0; cntr < len(srvc.Reals); cntr++ {
		go srvc.Reals[cntr].StartReal()
	}
	loop := 1
	srvc.ToAdapter <- GenerateAdapterMsg("AddService", srvc, nil)
	for loop == 1 {
		select {
		case msg := <-srvc.FromReal:
			switch msg.Cmnd {
			case "Dead":
				srvc.AliveReals--
				logMsg := strings.Join([]string{"real server", msg.Data, "now dead"}, " ")
				srvc.logWriter.Write([]byte(logMsg))
				rlSrv := srvc.FindRealFromMsgData(msg.Data)
				if rlSrv != nil {
					srvc.ToAdapter <- GenerateAdapterMsg("DeleteRealServer", srvc, rlSrv)
				}
				if srvc.AliveReals < srvc.Quorum && srvc.State == true {
					srvc.State = false
					logMsg = strings.Join([]string{"turning down service", srvc.VIP,
						srvc.Port, srvc.Proto}, " ")
					srvc.logWriter.Write([]byte(logMsg))
					srvc.ToAdapter <- GenerateAdapterMsg("WithdrawService", srvc, nil)
				}
			case "Alive":
				srvc.AliveReals++
				logMsg := strings.Join([]string{"real server", msg.Data, "now alive"}, " ")
				srvc.logWriter.Write([]byte(logMsg))
				rlSrv := srvc.FindRealFromMsgData(msg.Data)
				if rlSrv != nil {
					srvc.ToAdapter <- GenerateAdapterMsg("AddRealServer", srvc, rlSrv)
				}
				//TODO: hysteresis
				if srvc.AliveReals >= srvc.Quorum && srvc.State == false {
					srvc.State = true
					logMsg = strings.Join([]string{"bringing up service", srvc.VIP,
						srvc.Port, srvc.Proto}, " ")
					srvc.logWriter.Write([]byte(logMsg))
					srvc.ToAdapter <- GenerateAdapterMsg("AdvertiseService", srvc, nil)
				}
			case "RSFatalError":
				/*
				 right now it seems that real could send this tipe of msg only
				 at the beggining of his life (when it's not yet enbled and always counted
				 as dead; so we dont need to check if it's alive and do something with
				 AliveReals counter. Mb this would change in future
				*/
				DataFields := strings.Fields(msg.Data)
				if len(DataFields) < 3 {
					DataFields = append(DataFields, "")
				}
				rlSrv, index := srvc.FindReal(DataFields[0], DataFields[1], DataFields[2])
				if rlSrv != nil {
					srvc.RemoveReal(rlSrv, index, false)
				}
			}
		case msgFromSL := <-srvc.ToService:
			switch msgFromSL.Cmnd {
			case "GetInfo":
				data := strings.Join([]string{
					"State: ", strconv.FormatBool(srvc.State),
					"Alive Reals: ", strconv.Itoa(srvc.AliveReals),
					"Quorum: ", strconv.Itoa(srvc.Quorum)}, " ")
				srvc.FromService <- ServiceMsg{Data: data}
			case "RemoveService":
				if srvc.State {
					srvc.ToAdapter <- GenerateAdapterMsg("WithdrawService", srvc, nil)
				}
				for _, rlSrv := range srvc.Reals {
					rlSrv.ToReal <- ServiceMsg{Cmnd: "Shutdown"}
					//TODO: check if real was alive
					srvc.ToAdapter <- GenerateAdapterMsg("DeleteRealServer", srvc, &rlSrv)
				}
				//TODO: remove service from adapter
				logMsg := strings.Join([]string{"service ", srvc.VIP,
					srvc.Port, srvc.Proto, " successfully shuted down"}, " ")
				srvc.logWriter.Write([]byte(logMsg))
				loop = 0
			}
		}
	}
}

func (rlSrv *RealServer) ServiceMsgDataString() string {
	DataString := ""
	DataString = strings.Join([]string{rlSrv.RIP, rlSrv.Port, rlSrv.Meta}, " ")
	return DataString
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
			DataString := rlSrv.ServiceMsgDataString()
			rlSrv.ToService <- ServiceMsg{Cmnd: "RSFatalError", Data: DataString}
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
				DataString := rlSrv.ServiceMsgDataString()
				rlSrv.ToService <- ServiceMsg{Cmnd: "RSFatalError", Data: DataString}
				loop = 0
			case 1:
				if rlSrv.State == false {
					rlSrv.State = true
					DataString := rlSrv.ServiceMsgDataString()
					rlSrv.ToService <- ServiceMsg{Cmnd: "Alive", Data: DataString}
				}
			case 0:
				if rlSrv.State == true {
					rlSrv.State = false
					DataString := rlSrv.ServiceMsgDataString()
					rlSrv.ToService <- ServiceMsg{Cmnd: "Dead", Data: DataString}
				}
			}
		case msgToReal := <-rlSrv.ToReal:
			switch msgToReal.Cmnd {
			case "Shutdown":
				toCheck <- 1
				loop = 0
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
	if rlsrv.RIP == otherRlsrv.RIP && rlsrv.Port == otherRlsrv.Port &&
		rlsrv.Meta == otherRlsrv.Meta {
		return true
	} else {
		return false
	}
}

//TODO: add loging as well
func (srvc *Service) AddReal(rlsrv RealServer) int {
	for cntr := 0; cntr < len(srvc.Reals); cntr++ {
		if srvc.Reals[cntr].isEqualReal(rlsrv) {
			return -1
		}
	}
	rlsrv.Proto = srvc.Proto
	rlsrv.Timeout = srvc.Timeout
	rlsrv.ToService = srvc.FromReal
	rlsrv.ToReal = make(chan ServiceMsg)
	srvc.Reals = append(srvc.Reals, rlsrv)
	return len(srvc.Reals)

}

//TODO: add remove and change realServer logic
