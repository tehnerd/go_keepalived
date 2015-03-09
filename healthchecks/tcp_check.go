package healthchecks

import (
	"net"
	"strings"
	"time"
)

func TCPCheck(toCheck, fromCheck chan int, checkLine []string, timeOut int) {
	Addr := strings.Join(checkLine, ":")
	tcpAddr, err := net.ResolveTCPAddr("tcp", Addr)
	if err != nil {
		fromCheck <- -1
		return
	}
	loop := 1
	for loop == 1 {
		select {
		case <-toCheck:
			loop = 0
		case <-time.After(time.Second * time.Duration(timeOut)):
			/*
				TODO: add timeout for initial connection. if remote host silently drops
				our packets (for example iptables -j DROP) we will wait whole
				tcp to(exp bo)* tcp syn retries before fail this check
			*/
			tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
			if err != nil {
				fromCheck <- 0
			} else {
				fromCheck <- 1
				tcpConn.Close()
			}
		}
	}
}
