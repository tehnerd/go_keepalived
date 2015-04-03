package service

import (
	"strings"
)

func DispatchSLMsgs(sl *ServicesList) {
	for {
		select {
		case msg := <-sl.ToServiceList:
			responseStruct := make(map[string]string)
			switch msg.Cmnd {
			case "GetInfo":
				for _, service := range sl.List {
					service.ToService <- ServiceMsg{Cmnd: "GetInfo"}
					serviceResponse := <-service.FromService
					vip := strings.Join([]string{service.VIP, service.Proto, service.Port}, ":")
					responseStruct[vip] = serviceResponse.Data
				}
			}
			sl.FromServiceList <- ServiceMsg{DataMap: (&responseStruct)}
		}
	}
}
