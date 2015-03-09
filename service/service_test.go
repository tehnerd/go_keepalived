package service

import (
	"fmt"
	"go_keepalived/adapter"
	"log/syslog"
	"testing"
	"time"
)

func DummyAdapterReader(msgChan chan adapter.AdapterMsg) {
	for {
		msg := <-msgChan
		fmt.Println(msg)
	}
}

func TestServicesListAdd(t *testing.T) {
	srvc1 := Service{VIP: "127.0.0.1", Proto: "tcp", Port: "22"}
	srvc2 := Service{VIP: "127.0.0.2", Proto: "tcp", Port: "22"}
	srvc3 := Service{VIP: "127.0.0.3", Proto: "tcp", Port: "22"}
	srvc4 := Service{VIP: "127.0.0.1", Proto: "tcp", Port: "22"}
	srvcList := ServicesList{}
	srvcList.Init()
	//HACK. we wont need to send anything to ipvsadm
	srvcList.ToAdapter = make(chan adapter.AdapterMsg)
	go DummyAdapterReader(srvcList.ToAdapter)
	srvcList.Add(srvc1)
	srvcList.Add(srvc2)
	srvcList.Add(srvc3)
	srvcList.Add(srvc4)
	if len(srvcList.List) != 3 {
		t.Errorf("error in adding services")
	}
}

func TestServicesListRemove(t *testing.T) {
	srvc1 := Service{VIP: "127.0.0.1", Proto: "tcp", Port: "22"}
	srvc2 := Service{VIP: "127.0.0.2", Proto: "tcp", Port: "22"}
	srvc3 := Service{VIP: "127.0.0.3", Proto: "tcp", Port: "22"}
	srvc4 := Service{VIP: "127.0.0.4", Proto: "tcp", Port: "22"}
	srvcList := ServicesList{}
	srvcList.Init()
	srvcList.ToAdapter = make(chan adapter.AdapterMsg)
	go DummyAdapterReader(srvcList.ToAdapter)
	srvcList.Add(srvc1)
	srvcList.Add(srvc2)
	srvcList.Add(srvc3)
	srvcList.Add(srvc4)
	srvcList.Remove(srvc4)
	if len(srvcList.List) != 3 {
		t.Errorf("error in adding services")
	}
	srvcList.Remove(srvc2)
	if len(srvcList.List) != 2 {
		t.Errorf("error in adding services")
	}

}

func TestServiceAddReal(t *testing.T) {
	srvc1 := Service{VIP: "127.0.0.1", Proto: "tcp", Port: "22"}
	rlSrv1 := RealServer{RIP: "127.0.0.2", Port: "22"}
	rlSrv2 := RealServer{RIP: "127.0.0.3", Port: "22"}
	rlSrv3 := RealServer{RIP: "127.0.0.4", Port: "22"}
	rlSrv4 := RealServer{RIP: "127.0.0.4", Port: "22"}
	srvc1.Init()
	srvc1.ToAdapter = make(chan adapter.AdapterMsg)
	go DummyAdapterReader(srvc1.ToAdapter)
	srvc1.Timeout = 33
	srvc1.AddReal(rlSrv1)
	srvc1.AddReal(rlSrv2)
	srvc1.AddReal(rlSrv3)
	srvc1.AddReal(rlSrv4)

	if len(srvc1.Reals) != 3 {
		t.Errorf("error in adding realServers (duplicate detection logic)")
	}
	if srvc1.Reals[0].Timeout != 33 {
		t.Errorf("error in Timeout's val copying")
	}
	if srvc1.Reals[0].Proto != "tcp" {
		t.Errorf("error in Proto's val copying")
	}

}

func TestServiceRemoveReal(t *testing.T) {
	srvc1 := Service{VIP: "127.0.0.1", Proto: "tcp", Port: "22"}
	rlSrv1 := RealServer{RIP: "127.0.0.2", Port: "22"}
	rlSrv2 := RealServer{RIP: "127.0.0.3", Port: "22"}
	rlSrv3 := RealServer{RIP: "127.0.0.4", Port: "22"}
	srvc1.Init()
	//because we didnt add srvc1 to services list we dont have syslog Writer
	srvc1.logWriter, _ = syslog.New(syslog.LOG_INFO, "go_keepalive_test")
	srvc1.Timeout = 33
	srvc1.ToAdapter = make(chan adapter.AdapterMsg)
	go DummyAdapterReader(srvc1.ToAdapter)

	srvc1.AddReal(rlSrv1)
	srvc1.AddReal(rlSrv2)
	srvc1.AddReal(rlSrv3)
	_, index := srvc1.FindReal(rlSrv2.RIP, rlSrv2.Port, rlSrv2.Meta)
	if index != 1 {
		t.Errorf("FindReal is not working")
	}
	r, index := srvc1.FindReal(rlSrv3.RIP, rlSrv3.Port, rlSrv3.Meta)
	srvc1.RemoveReal(r, index, false)
	if len(srvc1.Reals) != 2 {
		t.Errorf("RemoveReal is not working")
	}
	srvc1.AddReal(rlSrv3)
	r, index = srvc1.FindReal(rlSrv2.RIP, rlSrv2.Port, rlSrv2.Meta)
	srvc1.RemoveReal(r, index, false)
	if len(srvc1.Reals) != 2 {
		t.Errorf("RemoveReal is not working")
	}

}

func TestServiceStart(t *testing.T) {
	srvc1 := Service{VIP: "127.0.0.1", Proto: "tcp", Port: "22"}
	rlSrv1 := RealServer{RIP: "127.0.0.1", Port: "60021", Check: "tcp"}
	rlSrv2 := RealServer{RIP: "127.0.0.1", Port: "60022", Check: "tcp"}
	rlSrv3 := RealServer{RIP: "127.0.0.1", Port: "600023", Check: "tcp"}
	rlSrv4 := RealServer{RIP: "127.0.0.1", Port: "60024", Check: "https"}
	srvc1.Init()
	srvc1.logWriter, _ = syslog.New(syslog.LOG_INFO, "go_keepalive_test")
	srvc1.Timeout = 33
	srvc1.ToAdapter = make(chan adapter.AdapterMsg)
	go DummyAdapterReader(srvc1.ToAdapter)
	srvc1.AddReal(rlSrv1)
	srvc1.AddReal(rlSrv2)
	srvc1.AddReal(rlSrv3)
	srvc1.AddReal(rlSrv4)
	go srvc1.StartService()
	//stupid way for sync, mb add something better in future
	time.Sleep(3 * time.Second)
	if len(srvc1.Reals) != 2 {
		t.Errorf("something wrong with RSFatalError handling, there are %d reals", len(srvc1.Reals))
	}
}
