package adapter

import (
	"fmt"
	"os/exec"
	"strings"
)

const (
	//TODO: mb add something else. like /usr/bin/ipvsadm etc
	IPVSADM_CMND = "ipvsadm"
)

/*
	This is ipvsadm's view of AdapterMsg struct
*/
type IPVSAdmCmnds struct {
	VIP string
	RIP string
	// "-A","-D","-a","-d", "-E", "-e"
	ActionFlag string
	// "-s"
	SchedFlag string
	// "-t", "-u", "-f"
	ProtoFlag string
	// "-w"
	WeightFlag string
	/*
	 flag, which indicates that command must contain realServers info
	 0 - means there is no realServer's info
	 1 - add or change
	 2 - delete
	*/
	RealServerComand int
	//method of lb; such as ipip or nat; -m or -i
	MetaInfo string
}

func IPVSParseAdapterMsg(msg *AdapterMsg) IPVSAdmCmnds {
	ipvsCmnds := IPVSAdmCmnds{}
	//TODO: check (prob regexp) for ipv6
	ipvsCmnds.VIP = strings.Join([]string{msg.ServiceVIP, msg.ServicePort}, ":")
	ipvsCmnds.RIP = strings.Join([]string{msg.RealServerRIP, msg.RealServerPort}, ":")
	switch msg.ServiceProto {
	case "tcp", "Tcp", "TCP":
		ipvsCmnds.ProtoFlag = "-t"
	case "udp", "UDP", "Udp":
		ipvsCmnds.ProtoFlag = "-u"
	case "fwmark", "FW":
		ipvsCmnds.ProtoFlag = "-f"
	}
	/*
		TODO: meta (masq or tun), weight, sched etc... now we are trying
		 to implement the most basic functions for POC
		metainfo hardcoded now just for the sake of implementing minimum working
		adapter
	*/
	ipvsCmnds.MetaInfo = "-m"
	return ipvsCmnds
}

func IPVSCliArgs(ipvsCmnds *IPVSAdmCmnds) []string {
	cliCmnds := []string{}
	cliCmnds = append(cliCmnds, ipvsCmnds.ActionFlag, ipvsCmnds.ProtoFlag,
		ipvsCmnds.VIP)
	if ipvsCmnds.RealServerComand > 0 {
		cliCmnds = append(cliCmnds, "-r", ipvsCmnds.RIP)
		if ipvsCmnds.RealServerComand == 1 {
			cliCmnds = append(cliCmnds, ipvsCmnds.MetaInfo)
		}
	}
	return cliCmnds
}

func IPVSAdmExec(msg *AdapterMsg) {
	ipvsCmnds := IPVSParseAdapterMsg(msg)
	switch msg.Type {
	case "AddService":
		ipvsCmnds.ActionFlag = "-A"
	case "DeleteService":
		ipvsCmnds.ActionFlag = "-D"
	case "ChangeService":
		ipvsCmnds.ActionFlag = "-E"
	case "AddRealServer":
		ipvsCmnds.ActionFlag = "-a"
		ipvsCmnds.RealServerComand = 1
	case "DeleteRealServer":
		ipvsCmnds.ActionFlag = "-d"
		ipvsCmnds.RealServerComand = 2
	case "ChangeRealServer":
		ipvsCmnds.ActionFlag = "-e"
		ipvsCmnds.RealServerComand = 1

	}
	cliArgs := IPVSCliArgs(&ipvsCmnds)
	fmt.Println(cliArgs)
	execCmnd := exec.Command(IPVSADM_CMND, cliArgs...)
	/*
		TODO: better check for error etc; sanitizing cliArgs (so we wont have "-A; rm -rf")
	*/
	output, err := execCmnd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(output))
}
