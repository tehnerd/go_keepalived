package api

import (
	"fmt"
	"go_keepalived/service"
	"regexp"
)

type APIMsg struct {
	Cmnd string
	Data *map[string]string
}

type GenericAPI struct {
	Enable          bool
	ToServiceList   chan service.ServiceMsg
	FromServiceList chan service.ServiceMsg
	ToApi           chan APIMsg
	FromApi         chan APIMsg
	HttpApi         bool
	//TODO:
	GRPCApi   bool
	ThriftApi bool
	//TODO: for authentication/authorization
	MasterPwd string
}

func InitAPI(api GenericAPI, ToServiceList, FromServiceList chan service.ServiceMsg) {
	api.ToApi = make(chan APIMsg)
	api.FromApi = make(chan APIMsg)
	api.ToServiceList = ToServiceList
	api.FromServiceList = FromServiceList
	if api.HttpApi {
		go StartHTTPApi(api.ToApi, api.FromApi)
	}
	for {
		select {
		case request := <-api.ToApi:
			resp := ProcessApiRequest(&api, request)
			if resp != nil {
				api.FromApi <- APIMsg{Data: &resp}
			} else {
				api.FromApi <- APIMsg{Data: &map[string]string{"result": "wrong command"}}
			}
		}
	}
}

/*
	Right now i want to implement this type of Command:
	"GetInfo" - get generic info about  services
	"AddReal" - add real server to service
	"RemoveReal" - remove real server from service
	"AddService" - add new service
	"RemoveService" - remove service
	"ChangeReal" - change data in real server (for example weight)
	"ChangeService" - change data in service (for example quorum etc)
	"AddPeer" - add new peer (fro bgp based notifier)
	"RemovePeer" - remove peer (fro bgp based notifier)
	"StopNotification" - stop advertise that service is available to external world
	"StartNotification" - start to advertise "alive" service to external world
*/

func ProcessApiRequest(api *GenericAPI, request APIMsg) map[string]string {
	//TODO: auth check
	cmnd := (*request.Data)["Command"]
	switch cmnd {
	case "GetInfo":
		api.ToServiceList <- service.ServiceMsg{Cmnd: "GetInfo"}
		resp := <-api.FromServiceList
		return (*resp.DataMap)
	case "AddService":
		err := serviceSanityCheck(request.Data)
		if err != nil {
			return nil
		}
		api.ToServiceList <- service.ServiceMsg{Cmnd: "AddService", DataMap: request.Data}
		resp := <-api.FromServiceList
		return (*resp.DataMap)
	case "RemoveService":
		err := serviceSanityCheck(request.Data)
		if err != nil {
			return nil
		}
		api.ToServiceList <- service.ServiceMsg{Cmnd: "RemoveService", DataMap: request.Data}
		resp := <-api.FromServiceList
		return (*resp.DataMap)
	case "ChangeService":
		err := serviceSanityCheck(request.Data)
		if err != nil {
			return nil
		}
		api.ToServiceList <- service.ServiceMsg{Cmnd: "ChangeService", DataMap: request.Data}
		resp := <-api.FromServiceList
		return (*resp.DataMap)
	case "AddReal":
		err := serviceSanityCheck(request.Data)
		if err != nil {
			return nil
		}
		err = realSrvSanityCheck(request.Data)
		if err != nil {
			return nil
		}
		api.ToServiceList <- service.ServiceMsg{Cmnd: "AddReal", DataMap: request.Data}
		resp := <-api.FromServiceList
		return (*resp.DataMap)
	case "RemoveReal":
		err := serviceSanityCheck(request.Data)
		if err != nil {
			return nil
		}
		err = realSrvSanityCheck(request.Data)
		if err != nil {
			return nil
		}
		api.ToServiceList <- service.ServiceMsg{Cmnd: "RemoveReal", DataMap: request.Data}
		resp := <-api.FromServiceList
		return (*resp.DataMap)
	case "ChangeReal":
		err := serviceSanityCheck(request.Data)
		if err != nil {
			return nil
		}
		err = realSrvSanityCheck(request.Data)
		if err != nil {
			return nil
		}
		api.ToServiceList <- service.ServiceMsg{Cmnd: "ChangeReal", DataMap: request.Data}
		resp := <-api.FromServiceList
		return (*resp.DataMap)
	case "AddPeer":
		err := peerSanityCheck(request.Data)
		if err != nil {
			return nil
		}
		api.ToServiceList <- service.ServiceMsg{Cmnd: "AddPeer", DataMap: request.Data}
		resp := <-api.FromServiceList
		return (*resp.DataMap)
	case "RemovePeer":
		err := peerSanityCheck(request.Data)
		if err != nil {
			return nil
		}
		api.ToServiceList <- service.ServiceMsg{Cmnd: "AddPeer", DataMap: request.Data}
		resp := <-api.FromServiceList
		return (*resp.DataMap)

	}
	return nil
}

func serviceSanityCheck(request *map[string]string) error {
	v4re, _ := regexp.Compile(`^(\d{1,3}\.){3}\d{1,3}$`)
	v6re, _ := regexp.Compile(`^\[((\d|a|b|c|d|e|f|A|B|C|D|E|F){0,4}\:?){1,8}\]$`)
	numRe, _ := regexp.Compile(`^\d+$`)
	if vip, exists := (*request)["VIP"]; !exists {
		return fmt.Errorf("request doesnt has mandatory VIP field\n")
	} else {
		if !v4re.MatchString(vip) && !v6re.MatchString(vip) {
			return fmt.Errorf("VIP is not v4 or v6 address\n")
		}
	}
	if _, exists := (*request)["Proto"]; !exists {
		return fmt.Errorf("request doesnt has mandatory Proto field\n")
	}
	if port, exists := (*request)["Port"]; !exists {
		return fmt.Errorf("request doesnt has mandatory Port field\n")
	} else if !numRe.MatchString(port) {
		return fmt.Errorf("port must be a number\n")
	}
	return nil
}

func realSrvSanityCheck(request *map[string]string) error {
	v4re, _ := regexp.Compile(`^(\d{1,3}\.){3}\d{1,3}$`)
	v6re, _ := regexp.Compile(`^\[((\d|a|b|c|d|e|f|A|B|C|D|E|F){0,4}\:?){1,8}\]$`)
	numRe, _ := regexp.Compile(`^\d+$`)
	if rip, exists := (*request)["RIP"]; !exists {
		return fmt.Errorf("request doesnt has mandatory RIP field\n")
	} else {
		if !v4re.MatchString(rip) && !v6re.MatchString(rip) {
			return fmt.Errorf("RIP is not v4 or v6 address\n")
		}
	}
	if _, exists := (*request)["Check"]; !exists {
		return fmt.Errorf("request doesnt has mandatory Check field\n")
	}
	if port, exists := (*request)["RealPort"]; !exists {
		return fmt.Errorf("request doesnt has mandatory Port field\n")
	} else if !numRe.MatchString(port) {
		return fmt.Errorf("port must be a number\n")
	}
	return nil
}

func peerSanityCheck(request *map[string]string) error {
	v4re, _ := regexp.Compile(`^(\d{1,3}\.){3}\d{1,3}$`)
	v6re, _ := regexp.Compile(`^((\d|a|b|c|d|e|f|A|B|C|D|E|F){0,4}\:?){1,8}$`)
	if addr, exists := (*request)["Address"]; !exists {
		return fmt.Errorf("request doesnt has mandatory Address field\n")
	} else {
		if !v4re.MatchString(addr) && !v6re.MatchString(addr) {
			return fmt.Errorf("Address is not v4 or v6 address\n")
		}
	}
	return nil
}
