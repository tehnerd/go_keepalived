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
					vip := createVipName(service)
					responseStruct[vip] = serviceResponse.Data
				}
			case "AddService":
				srvc := serviceFromDefinition(msg.DataMap)
				err := sl.Add(srvc)
				if err != nil {
					break
				}
				srvc = sl.List[len(sl.List)-1]
				go srvc.StartService()
				vip := createVipName(srvc)
				responseStruct["result"] = "Successfully added new service:" + vip
			case "RemoveService":
				srvc := serviceFromDefinition(msg.DataMap)
				err := sl.Remove(srvc)
				if err != nil {
					break
				}
				vip := createVipName(srvc)
				responseStruct["result"] = "Successfully removed service:" + vip
			}
			sl.FromServiceList <- ServiceMsg{DataMap: (&responseStruct)}
		}
	}
}

func serviceFromDefinition(definition *map[string]string) Service {
	srvc := Service{}
	srvc.Init()
	srvc.VIP = (*definition)["VIP"]
	srvc.Proto = (*definition)["Proto"]
	srvc.Port = (*definition)["Port"]
	return srvc
}

func createVipName(srvc Service) string {
	return strings.Join([]string{srvc.VIP, srvc.Proto, srvc.Port}, ":")
}
