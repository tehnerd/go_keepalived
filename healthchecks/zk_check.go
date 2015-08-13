package healthchecks

import (
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

func ZKCheck(toCheck, fromCheck chan int, checkLine []string, timeOut int) {
	zkAddr := checkLine[0]
	zkPath := checkLine[1]
	conn, connEvent, err := zk.Connect([]string{zkAddr}, time.Duration(timeOut)*time.Second)
	if err != nil {
	CONNECT_LOOP:
		for {
			conn, connEvent, err = zk.Connect([]string{zkAddr}, time.Duration(timeOut)*time.Second)
			if err != nil {
				select {
				case <-time.After(time.Duration(timeOut) * time.Second):
				case <-toCheck:
					return
				}
			} else {
				break CONNECT_LOOP
			}
		}
	}
CHECK_LOOP:
	for {
		exists, _, event, err := conn.ExistsW(zkPath)
		if err != nil {
			select {
			case fromCheck <- 0:
			case <-toCheck:
				break CHECK_LOOP
			}
			select {
			case <-time.After(time.Second * 10):
				continue CHECK_LOOP
			case <-toCheck:
				break CHECK_LOOP
			}
		}
		if exists {
			select {
			case fromCheck <- 1:
			case <-toCheck:
				break CHECK_LOOP
			}
		} else {
			select {
			case fromCheck <- 0:
			case <-toCheck:
				break CHECK_LOOP
			}
		}
		select {
		case <-connEvent:
		case <-toCheck:
			break CHECK_LOOP
		case <-event:
		}
	}
}
