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
				fromCheck <- 0
			} else if resp.StatusCode != http.StatusOK {
				fromCheck <- 0
				resp.Body.Close()
			} else {
				fromCheck <- 1
				resp.Body.Close()
			}
		}
	}
}
