package adapter

import (
	"go_keepalived/notifier"
)

/*
 this type of messages are using to communicate to/from
 adapter (right now we are using only ipvsadm as adapter for
 datapath; mb will add something else in the future (i hope there would be
 libnl for go; so we wont need external bin's). we are trying to be implementation
 agnostic (so it would be easy to add something else (for example l7 slb's cfg parser etc))
*/
type AdapterMsg struct {
	/*
		Message Type. Could be:
		"AddX","ChangeX","DeleteX". msgs derection Service->Adapter; where X could be either
			"Service" or "Real"
		"Error", Adapter->Service
		"AdvertiseService","WithdrawService" when service goes up/down
	*/
	Type string

	/*
		We need to be able to distinguish one service from another. service could be
		uniquely identified by: VIP,Proto,Port
	*/
	ServiceVIP   string
	ServicePort  string
	ServiceProto string

	/*
		Same as for service, but for realServers. Could be identified by
		RIP,Port,Meta. Also we want to be able to change Reals weight (in future)
		So we are passing Real's weight as well
	*/
	RealServerRIP    string
	RealServerPort   string
	RealServerMeta   string
	RealServerWeight string
}

func StartAdapter(msgChan chan AdapterMsg, notifierChan chan notifier.NotifierMsg) {
	loop := 1
	for loop == 1 {
		select {
		case msg := <-msgChan:
			/*
				TODO: if we ever gonna have more than ipvs adapter we need to add here
				adapter's type check
			*/
			err := IPVSAdmExec(&msg)
			if err != nil {
				//TODO: proper handling
				continue
			}
			/*
				The whole purpose of notifier is to tell about state of our services to
				outerworld (right now the only implemented notifier is bgp injector)
			*/
			notifierChan <- notifier.NotifierMsg{Type: msg.Type, Data: msg.ServiceVIP}
		}
	}
}
