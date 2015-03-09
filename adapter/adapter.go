package adapter

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

func StartAdapter(msgChan chan AdapterMsg) {
	loop := 1
	for loop == 1 {
		select {
		case msg := <-msgChan:
			/*
				TODO: if we ever gonna have more than ipvs adapter we need to add here
				adapter's type check
			*/
			IPVSAdmExec(&msg)
		}
	}
}
