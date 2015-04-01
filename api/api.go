package api

import (
	"go_keepalived/service"
)

type APIMsg struct {
	Cmnd string
	Data []byte
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
		case <-api.ToApi:
		case <-api.FromServiceList:
		}
	}
}
