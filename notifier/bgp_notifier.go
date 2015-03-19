package notifier

import (
	"github.com/tehnerd/bgp"
	"strings"
)

func BGPNotifier(msgChan chan NotifierMsg, responseChan chan NotifierMsg,
	notifierConfig NotifierConfig) {
	bgpMainContext := bgp.BGPContext{ASN: notifierConfig.ASN, ListenLocal: notifierConfig.ListenLocal}
	toBGPProcess := make(chan bgp.BGPProcessMsg)
	fromBGPProcess := make(chan bgp.BGPProcessMsg)
	go bgp.StartBGPProcess(toBGPProcess, fromBGPProcess, bgpMainContext)
	/*
		table of service's refcount. we could has two different services on the same vip (for example tcp/80; tcp/443)
		//TODO: overflow check
	*/
	serviceTable := make(map[string]uint32)
	for {
		msg := <-msgChan
		switch msg.Type {
		case "AdvertiseService":
			if _, exists := serviceTable[msg.Data]; exists {
				serviceTable[msg.Data]++
				//more that one service has same vip; we have already advertised it
				if serviceTable[msg.Data] > 1 {
					continue
				}
			} else {
				serviceTable[msg.Data] = 1
			}
			//TODO: check/parse if route v4 or v6
			//we advertise or withdraw only host routes
			Route := strings.Join([]string{msg.Data, "/32"}, "")
			toBGPProcess <- bgp.BGPProcessMsg{
				Cmnd: "AddV4Route",
				Data: Route}
		case "WithdrawService":
			if _, exists := serviceTable[msg.Data]; exists {
				if serviceTable[msg.Data] == 0 {
					continue
				}
				serviceTable[msg.Data]--
				//there are still alive services who uses this vip
				if serviceTable[msg.Data] > 0 {
					continue
				}
			} else {
				continue
			}
			//TODO: check/parse if route v4 or v6
			//we advertise or withdraw only host routes
			Route := strings.Join([]string{msg.Data, "/32"}, "")
			toBGPProcess <- bgp.BGPProcessMsg{
				Cmnd: "WithdrawV4Route",
				Data: Route}
		case "AddPeer":
			//TODO: check/parse v4/v6
			toBGPProcess <- bgp.BGPProcessMsg{
				Cmnd: "AddNeighbour",
				Data: msg.Data}
			//TODO: RemovePeer here and in simple_bgp_injector
		}

	}
}
