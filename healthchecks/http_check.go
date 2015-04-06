package healthchecks

import (
	"net/http"
	"time"
)

func HTTPCheck(toCheck, fromCheck chan int, checkLine []string, timeOut int) {
	if len(checkLine) < 1 {
		fromCheck <- -1
		return
	}
	client := &http.Client{Timeout: time.Duration(timeOut)}
	loop := 1
	for loop == 1 {
		select {
		case <-toCheck:
			loop = 0
		case <-time.After(time.Second * time.Duration(timeOut)):
			resp, err := client.Get(checkLine[0])
			if err != nil {
				select {
				case fromCheck <- 0:
				case <-toCheck:
					loop = 0
				}
			} else if resp.StatusCode != http.StatusOK {
				select {
				case fromCheck <- 0:
					resp.Body.Close()
				case <-toCheck:
					loop = 0
				}
			} else {
				select {
				case fromCheck <- 1:
					resp.Body.Close()
				case <-toCheck:
					loop = 0
				}
			}
		}
	}
}
