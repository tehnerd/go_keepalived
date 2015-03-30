package notifier

import (
	"github.com/tehnerd/bgp2go"
	"regexp"
	"strings"
)

func BGPNotifier(msgChan chan NotifierMsg, responseChan chan NotifierMsg,
	notifierConfig NotifierConfig) {
	v4re, _ := regexp.Compile(`^(\d{1,3}\.){3}\d{1,3}$`)
	v6re, _ := regexp.Compile(`^\[(((\d|a|b|c|d|e|f|A|B|C|D|E|F){0,4}\:?){1,8})\]$`)
	bgpMainContext := bgp2go.BGPContext{ASN: notifierConfig.ASN, ListenLocal: notifierConfig.ListenLocal}
	toBGPProcess := make(chan bgp2go.BGPProcessMsg)
	fromBGPProcess := make(chan bgp2go.BGPProcessMsg)
	go bgp2go.StartBGPProcess(toBGPProcess, fromBGPProcess, bgpMainContext)
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
			if v4re.MatchString(msg.Data) {
				Route := strings.Join([]string{msg.Data, "/32"}, "")
				toBGPProcess <- bgp2go.BGPProcessMsg{
					Cmnd: "AddV4Route",
					Data: Route}
			} else {
				//TODO: mb use regexp findstring instead
				prefix := v6re.FindStringSubmatch(msg.Data)
				if len(prefix) >= 2 {
					Route := strings.Join([]string{prefix[1], "/128"}, "")
					toBGPProcess <- bgp2go.BGPProcessMsg{
						Cmnd: "AddV6Route",
						Data: Route}
				}
			}
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

			if v4re.MatchString(msg.Data) {
				Route := strings.Join([]string{msg.Data, "/32"}, "")
				toBGPProcess <- bgp2go.BGPProcessMsg{
					Cmnd: "WithdrawV4Route",
					Data: Route}
			} else {
				prefix := v6re.FindStringSubmatch(msg.Data)
				if len(prefix) >= 2 {

					Route := strings.Join([]string{prefix[1], "/128"}, "")
					toBGPProcess <- bgp2go.BGPProcessMsg{
						Cmnd: "WithdrawV6Route",
						Data: Route}
				}
			}
		case "AddPeer":
			if v4re.MatchString(msg.Data) {
				toBGPProcess <- bgp2go.BGPProcessMsg{
					Cmnd: "AddNeighbour",
					Data: msg.Data}
			} else {
				//TODO: sanity check for v6 address
				//for ResolveTCPAddr to work ipv6 must be in format [<addr>]
				data := strings.Join([]string{"[", msg.Data, "]"}, "")
				data = strings.Join([]string{data, "inet6"}, " ")
				toBGPProcess <- bgp2go.BGPProcessMsg{
					Cmnd: "AddNeighbour",
					Data: data}
			}
			//TODO: RemovePeer here and in simple_bgp_injector
		}

	}
}
