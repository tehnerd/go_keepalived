package healthchecks

import (
	"net"
	"strings"
	"time"
)

func TCPCheck(toCheck, fromCheck chan int, checkLine []string, timeOut int) {
	Addr := strings.Join(checkLine, ":")
	loop := 1
	for loop == 1 {
		select {
		case <-toCheck:
			loop = 0
		case <-time.After(time.Second * time.Duration(timeOut)):
			tcpConn, err := net.DialTimeout("tcp", Addr,
				time.Second*time.Duration(timeOut))
			if err != nil {
				select {
				case fromCheck <- 0:
				case <-toCheck:
					loop = 0
				}
			} else {
				select {
				case fromCheck <- 1:
					tcpConn.Close()
				case <-toCheck:
					loop = 0
				}
			}
		}
	}
}
