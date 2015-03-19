package cfgparser

import (
	"bufio"
	"fmt"
	"go_keepalived/notifier"
	"go_keepalived/service"
	"os"
	"strconv"
	"strings"
)

//TODO: unittests
func ReadCfg(cfgFile string) (*service.ServicesList, *notifier.NotifierConfig) {
	fd, err := os.Open(cfgFile)
	defer fd.Close()
	if err != nil {
		fmt.Println("cant open cfg file")
		os.Exit(-1)
	}
	sl := service.ServicesList{}
	nc := notifier.NotifierConfig{}
	sl.Init()
	srvc := service.Service{}
	srvc.Init()
	rlSrv := service.RealServer{}
	/*
		counting "{"; so we will be able to see when old servicse defenition stops and new one starts
	*/
	sectionCntr := 0
	generalCfg := false
	serviceCfg := false
	notifierCfg := false

	scanner := bufio.NewScanner(fd)
	/* main section of cfg parsing */
	line := 0
	for scanner.Scan() {
		line++
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			continue
		}
		if fields[0] == "service" {
			if len(fields) < 4 {
				fmt.Println("line: ", line)
				fmt.Println("error in parsing service's defenition\n must be service <vip> <port> {")
				os.Exit(-1)
			}
			if sectionCntr != 0 {
				fmt.Println("line: ", line)
				fmt.Println("error in cfg file. it seems that you defined service inside other service")
				fmt.Println(fields)
				os.Exit(-1)
			}
			if fields[3] != "{" {
				fmt.Println("line: ", line)
				fmt.Println("error in parsing service's defenition\n must be service <vip> <port> {")
				os.Exit(-1)
			}
			serviceCfg = true
			sectionCntr++
			srvc.VIP = fields[1]
			srvc.Port = fields[2]
		} else if fields[0] == "real" && serviceCfg == true {
			if len(fields) < 4 {
				fmt.Println("line: ", line)
				fmt.Println("error in parsing real's defenition\n must be real <rip> <port> {")
				os.Exit(-1)
			}
			if sectionCntr != 1 {
				fmt.Println("line: ", line)
				fmt.Println("error in cfg file. it seems that you defined real inside other real or real outside of service")
				fmt.Println(fields)
				os.Exit(-1)
			}
			if fields[3] != "{" {
				fmt.Println("line: ", line)
				fmt.Println("error in parsing real's defenition\n must be real <rip> <port> {")
				os.Exit(-1)
			}
			sectionCntr++
			rlSrv.RIP = fields[1]
			rlSrv.Port = fields[2]
		} else if fields[0] == "general" {
			if sectionCntr != 0 || (len(fields) != 2 && fields[1] != "{") {
				fmt.Println("line: ", line)
				fmt.Println("error in parsing genaral's cfg defenition\n must be general {")
				os.Exit(-1)

			}
			generalCfg = true
			sectionCntr++
		} else if fields[0] == "notifier" {
			if sectionCntr != 1 || (len(fields) != 3 && fields[2] != "{") || generalCfg != true {
				fmt.Println("line: ", line)
				fmt.Println("error in parsing notifiers's cfg defenition\n must be notifier <type> {")
				os.Exit(-1)
			}
			nc.Type = fields[1]
			notifierCfg = true
			sectionCntr++
		} else if fields[0] == "}" {
			if sectionCntr == 2 {
				if serviceCfg == true {
					srvc.AddReal(rlSrv)
					rlSrv = service.RealServer{}
				}
				if notifierCfg == true {
					notifierCfg = false
				}
				sectionCntr--
			} else if sectionCntr == 1 {
				if serviceCfg == true {
					if service.IsServiceValid(srvc) {
						sl.Add(srvc)

					} else {
						fmt.Println("unvalid service defenition, fields VIP,PORT and Proto are mandatory")
					}
					srvc = service.Service{}
					srvc.Init()
					serviceCfg = false
				}
				if generalCfg == true {
					generalCfg = false
				}
				sectionCntr--
			} else {
				fmt.Println("line: ", line)
				fmt.Println("error in section's count (not enough/too many {})")
				os.Exit(-1)
			}
		} else if len(fields) > 1 {
			if serviceCfg == true {
				if fields[0] == "proto" && sectionCntr == 1 {
					srvc.Proto = fields[1]
				} else if fields[0] == "quorum" && sectionCntr == 1 {
					qnum, err := strconv.Atoi(fields[1])
					if err != nil {
						fmt.Println("line: ", line)
						fmt.Println("cant convert quorum to int")
						os.Exit(-1)
					}
					srvc.Quorum = qnum
				} else if fields[0] == "timeout" && sectionCntr == 1 {
					timeout, err := strconv.Atoi(fields[1])
					if err != nil {
						fmt.Println("line: ", line)
						fmt.Println("cant convert timeout to int")
						os.Exit(-1)
					}
					srvc.Timeout = timeout
				} else if fields[0] == "hysteresis" && sectionCntr == 1 {
					hnum, err := strconv.Atoi(fields[1])
					if err != nil {
						fmt.Println("line: ", line)
						fmt.Println("cant convert hysteresis to int")
						os.Exit(-1)
					}
					srvc.Hysteresis = hnum
				} else if fields[0] == "check" && sectionCntr == 2 {
					check := strings.Join(fields[1:], " ")
					rlSrv.Check = check
				} else if fields[0] == "meta" && sectionCntr == 2 {
					meta := strings.Join(fields[1:], " ")
					rlSrv.Meta = meta
				}
			} else if notifierCfg == true {
				if fields[0] == "ASN" {
					val, err := strconv.ParseUint(fields[1], 10, 32)
					if err != nil {
						fmt.Println("line: ", line)
						fmt.Println("cant convert asn to uint32")
						os.Exit(-1)
					}
					nc.ASN = uint32(val)
				} else if fields[0] == "listen" {
					if fields[1] == "enable" {
						nc.ListenLocal = true
					}
				} else if fields[0] == "neighbour" {
					nc.NeighboursList = append(nc.NeighboursList, fields[1])
				}
			}
		}
	}
	sl.AddNotifier(nc)
	return &sl, &nc
}
