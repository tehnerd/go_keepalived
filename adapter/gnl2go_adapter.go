package adapter

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/tehnerd/gnl2go"
)

const (
	AF_INET6 = 10
	AF_INET  = 2
)

var (
	supportedCommands = createMap([]string{
		"AddService",
		"DeleteService",
		"ChangeService",
		"AddRealServer",
		"DeleteRealServer",
		"ChangeRealServer",
	})
)

func createMap(stringSlice []string) map[string]bool {
	m := make(map[string]bool)
	for _, v := range stringSlice {
		m[v] = true
	}
	return m
}

type NetlinkCommand struct {
	Command   string
	Vip       string
	Rip       string
	Scheduler string
	VipPort   uint16
	RipPort   uint16
	Protocol  uint16
	Weight    int32
	Fwd       uint32
	VAF       uint16
	FWMark    uint32
}

func parseAdapterMsg(msg *AdapterMsg) (NetlinkCommand, error) {
	v6re, _ := regexp.Compile(`^\[(((\d|a|b|c|d|e|f|A|B|C|D|E|F){0,4}\:?){1,8})\]$`)
	var nlcmnd NetlinkCommand
	nlcmnd.Command = msg.Type
	if v6re.MatchString(msg.ServiceVIP) {
		vip := v6re.FindStringSubmatch(msg.ServiceVIP)
		nlcmnd.Vip = vip[1]
	} else {
		nlcmnd.Vip = msg.ServiceVIP
	}
	if v6re.MatchString(msg.RealServerRIP) {
		rip := v6re.FindStringSubmatch(msg.RealServerRIP)
		nlcmnd.Rip = rip[1]
	} else {
		nlcmnd.Rip = msg.RealServerRIP
	}
	if len(msg.ServicePort) > 0 {
		vport, err := strconv.ParseUint(msg.ServicePort, 10, 16)
		if err != nil {
			return nlcmnd, err
		}
		nlcmnd.VipPort = uint16(vport)
	}
	if len(msg.RealServerPort) > 0 {
		rport, err := strconv.ParseUint(msg.RealServerPort, 10, 16)
		if err != nil {
			return nlcmnd, err
		}
		nlcmnd.RipPort = uint16(rport)
	}
	nlcmnd.Protocol = uint16(gnl2go.ToProtoNum(gnl2go.NulStringType(msg.ServiceProto)))
	/* this means that service isn't tcp or udp - theorefor fwmark */
	if nlcmnd.Protocol == 0 {
		fwmark, err := strconv.ParseUint(msg.ServiceVIP, 10, 32)
		if err != nil {
			return nlcmnd, err
		}
		nlcmnd.FWMark = uint32(fwmark)
		nlcmnd.Vip = "fwmark"
		nlcmnd.VipPort = 0
		nlcmnd.RipPort = 0
		nlcmnd.VAF = AF_INET
	}

	if len(msg.RealServerWeight) > 0 {
		weight, err := strconv.ParseInt(msg.RealServerWeight, 10, 32)
		if err != nil {
			return nlcmnd, err
		}
		nlcmnd.Weight = int32(weight)
		/* srvces metainfo will contain: scheduler vip_address_family etc */
	}
	srvcMetaFields := strings.Fields(msg.ServiceMeta)
	if len(srvcMetaFields) >= 1 {
		nlcmnd.Scheduler = srvcMetaFields[0]
	} else {
		nlcmnd.Scheduler = "wrr"
	}
	if len(srvcMetaFields) >= 2 {
		switch srvcMetaFields[1] {
		case "ipv6":
			nlcmnd.VAF = AF_INET6
		default:
			nlcmnd.VAF = AF_INET
		}
	}
	realMetaFields := strings.Fields(msg.RealServerMeta)
	if len(realMetaFields) >= 1 {
		switch msg.RealServerMeta {
		case "tunnel", "tun":
			nlcmnd.Fwd = gnl2go.IPVS_TUNNELING
		default:
			nlcmnd.Fwd = gnl2go.IPVS_MASQUERADING
		}
	} else {
		nlcmnd.Fwd = gnl2go.IPVS_MASQUERADING
	}
	return nlcmnd, nil
}

func ExecCmnd(nlcmnd NetlinkCommand) error {
	ipvs := new(gnl2go.IpvsClient)
	err := ipvs.Init()
	if err != nil {
		return err
	}
	defer ipvs.Exit()
	switch nlcmnd.Command {
	case "AddService":
		if nlcmnd.Vip == "fwmark" {
			return ipvs.AddFWMService(nlcmnd.FWMark,
				nlcmnd.Scheduler, nlcmnd.VAF)
		} else {
			return ipvs.AddService(nlcmnd.Vip, nlcmnd.VipPort,
				nlcmnd.Protocol, nlcmnd.Scheduler)
		}
	case "DeleteService":
		if nlcmnd.Vip == "fwmark" {
			return ipvs.DelFWMService(nlcmnd.FWMark, nlcmnd.VAF)
		} else {
			return ipvs.DelService(nlcmnd.Vip, nlcmnd.VipPort, nlcmnd.Protocol)
		}
	case "ChangeService":
		return nil
	case "AddRealServer":
		if nlcmnd.Vip == "fwmark" {
			return ipvs.AddFWMDestFWD(nlcmnd.FWMark, nlcmnd.Rip, nlcmnd.VAF,
				nlcmnd.VipPort, nlcmnd.Weight, nlcmnd.Fwd)
		} else {
			return ipvs.AddDestPort(nlcmnd.Vip, nlcmnd.VipPort, nlcmnd.Rip,
				nlcmnd.RipPort, nlcmnd.Protocol, nlcmnd.Weight, nlcmnd.Fwd)
		}
	case "DeleteRealServer":
		if nlcmnd.Vip == "fwmark" {
			return ipvs.DelFWMDest(nlcmnd.FWMark, nlcmnd.Rip, nlcmnd.VAF,
				nlcmnd.RipPort)
		} else {
			return ipvs.DelDestPort(nlcmnd.Vip, nlcmnd.VipPort, nlcmnd.Rip,
				nlcmnd.RipPort, nlcmnd.Protocol)
		}
	case "ChangeRealServer":
		if nlcmnd.Vip == "fwmark" {
			return ipvs.UpdateFWMDestFWD(nlcmnd.FWMark, nlcmnd.Rip, nlcmnd.VAF,
				nlcmnd.RipPort, nlcmnd.Weight, nlcmnd.Fwd)
		} else {
			return ipvs.UpdateDestPort(nlcmnd.Vip, nlcmnd.VipPort, nlcmnd.Rip,
				nlcmnd.RipPort, nlcmnd.Protocol, nlcmnd.Weight, nlcmnd.Fwd)
		}
	}
	return nil
}

func GNLExec(msg *AdapterMsg) error {
	if _, exists := supportedCommands[msg.Type]; !exists {
		return nil
	}
	nlcmnd, err := parseAdapterMsg(msg)
	if err != nil {
		return err
	}
	return ExecCmnd(nlcmnd)
}
