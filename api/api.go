package api

import (
	"go_keepalived/service"
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
		case <-api.FromServiceList:
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
	}
	return nil
}
