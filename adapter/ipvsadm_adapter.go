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
	// "-s"; select schedulers type
	SchedFlag string
	// "-t", "-u", "-f"
	ProtoFlag string
	// "-w"
	WeightFlag string
	/*
	 flag, which indicates that command must include/exclude particular fields
	 0 - means there is no realServer's info and we adding service
	 1 - delete service
	 128 - add or change real
	 129 - delete real
	*/
	CommandFlag uint8
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
	ipvsCmnds.MetaInfo = msg.RealServerMeta
	ipvsCmnds.SchedFlag = msg.ServiceMeta
	ipvsCmnds.WeightFlag = msg.RealServerWeight
	return ipvsCmnds
}

func IPVSCliArgs(ipvsCmnds *IPVSAdmCmnds) []string {
	cliCmnds := []string{}
	cliCmnds = append(cliCmnds, ipvsCmnds.ActionFlag, ipvsCmnds.ProtoFlag,
		ipvsCmnds.VIP)
	if ipvsCmnds.CommandFlag > 127 {
		cliCmnds = append(cliCmnds, "-r", ipvsCmnds.RIP)
		if ipvsCmnds.CommandFlag == 128 {
			switch ipvsCmnds.MetaInfo {
			case "tunnel":
				cliCmnds = append(cliCmnds, "-i")
			default:
				cliCmnds = append(cliCmnds, "-m")
			}
			// this section is for add/change  real commands
			cliCmnds = append(cliCmnds, "-w", ipvsCmnds.WeightFlag)
		}
	} else {
		switch ipvsCmnds.CommandFlag {
		case 0:
			cliCmnds = append(cliCmnds, "-s", ipvsCmnds.SchedFlag)
		}
	}
	return cliCmnds
}

func IPVSAdmExec(msg *AdapterMsg) error {
	ipvsCmnds := IPVSParseAdapterMsg(msg)
	switch msg.Type {
	case "AdvertiseService", "WithdrawService":
		return nil
	case "AddService":
		ipvsCmnds.ActionFlag = "-A"
	case "DeleteService":
		ipvsCmnds.ActionFlag = "-D"
		ipvsCmnds.CommandFlag = 1
	case "ChangeService":
		ipvsCmnds.ActionFlag = "-E"
	case "AddRealServer":
		ipvsCmnds.ActionFlag = "-a"
		ipvsCmnds.CommandFlag = 128
	case "DeleteRealServer":
		ipvsCmnds.ActionFlag = "-d"
		ipvsCmnds.CommandFlag = 129
	case "ChangeRealServer":
		ipvsCmnds.ActionFlag = "-e"
		ipvsCmnds.CommandFlag = 128

	}
	cliArgs := IPVSCliArgs(&ipvsCmnds)
	fmt.Println(cliArgs)
	execCmnd := exec.Command(IPVSADM_CMND, cliArgs...)
	/*
		TODO: better check for error etc; sanitizing cliArgs (so we wont have "-A; rm -rf")
	*/
	output, err := execCmnd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}
