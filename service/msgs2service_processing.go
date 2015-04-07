package service

import (
	"strconv"
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
				responseStruct["result"] = "true"
				responseStruct["info"] = "Successfully added new service:" + vip
			case "RemoveService":
				srvc := serviceFromDefinition(msg.DataMap)
				err := sl.Remove(srvc)
				if err != nil {
					break
				}
				vip := createVipName(srvc)
				responseStruct["result"] = "true"
				responseStruct["info"] = "Successfully removed service:" + vip
			case "AddReal":
				srvcName := serviceFromDefinition(msg.DataMap)
				rlSrv := realSrvFromDefinition(msg.DataMap)
				i := sl.FindService(&srvcName)
				if i == -1 {
					break
				}
				sl.List[i].ToService <- ServiceMsg{Cmnd: "AddReal", DataMap: msg.DataMap}
				result := <-sl.List[i].FromService
				if result.Data == "RealAdded" {
					vip := createVipName(sl.List[i])
					rip := createRlSrvName(rlSrv)
					responseStruct["result"] = "true"
					responseStruct["info"] = "Successfully added new real " + rip + " for service:" + vip
				} else {
					responseStruct["result"] = "false"
				}
			case "RemoveReal":
				srvcName := serviceFromDefinition(msg.DataMap)
				rlSrv := realSrvFromDefinition(msg.DataMap)
				i := sl.FindService(&srvcName)
				if i == -1 {
					break
				}
				sl.List[i].ToService <- ServiceMsg{Cmnd: "RemoveReal",
					Data: strings.Join([]string{rlSrv.RIP, rlSrv.Port, rlSrv.Meta}, " ")}
				result := <-sl.List[i].FromService
				if result.Data == "RealRemoved" {
					vip := createVipName(sl.List[i])
					rip := createRlSrvName(rlSrv)
					responseStruct["result"] = "true"
					responseStruct["info"] = "Successfully removed real " + rip + " for service:" + vip
				} else {
					responseStruct["result"] = "false"
				}
			}
			sl.FromServiceList <- ServiceMsg{DataMap: (&responseStruct)}
		}
	}
}

//TODO: scheduler, lbmethod (nat or tun) , timeout, etc
func serviceFromDefinition(definition *map[string]string) Service {
	srvc := Service{}
	srvc.Init()
	srvc.VIP = (*definition)["VIP"]
	srvc.Proto = (*definition)["Proto"]
	srvc.Port = (*definition)["Port"]
	if quorum, exists := (*definition)["Quorum"]; exists {
		q, err := strconv.Atoi(quorum)
		if err != nil {
			srvc.Quorum = 1
		} else {
			srvc.Quorum = q
		}

	} else {
		srvc.Quorum = 1
	}
	return srvc
}

//TODO: meta. weight etc
func realSrvFromDefinition(definition *map[string]string) RealServer {
	rlSrv := RealServer{}
	rlSrv.RIP = (*definition)["RIP"]
	rlSrv.Port = (*definition)["RealPort"]
	rlSrv.Check = (*definition)["Check"]
	return rlSrv
}

func createVipName(srvc Service) string {
	return strings.Join([]string{srvc.VIP, srvc.Proto, srvc.Port}, ":")
}

func createRlSrvName(rlSrv RealServer) string {
	return strings.Join([]string{rlSrv.RIP, rlSrv.Proto, rlSrv.Port}, ":")
}
