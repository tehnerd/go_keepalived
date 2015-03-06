package service

import (
	"testing"
)

func TestServicesListAdd(t *testing.T) {
	srvc1 := Service{VIP: "1.1.1.1", Proto: "tcp", Port: "22"}
	srvc2 := Service{VIP: "1.1.1.2", Proto: "tcp", Port: "22"}
	srvc3 := Service{VIP: "1.1.1.3", Proto: "tcp", Port: "22"}
	srvc4 := Service{VIP: "1.1.1.1", Proto: "tcp", Port: "22"}
	srvcList := ServicesList{}
	srvcList.Init()
	srvcList.Add(srvc1)
	srvcList.Add(srvc2)
	srvcList.Add(srvc3)
	srvcList.Add(srvc4)
	if len(srvcList.List) != 3 {
		t.Errorf("error in adding services")
	}
}

func TestServicesListRemove(t *testing.T) {
	srvc1 := Service{VIP: "1.1.1.1", Proto: "tcp", Port: "22"}
	srvc2 := Service{VIP: "1.1.1.2", Proto: "tcp", Port: "22"}
	srvc3 := Service{VIP: "1.1.1.3", Proto: "tcp", Port: "22"}
	srvc4 := Service{VIP: "1.1.1.4", Proto: "tcp", Port: "22"}
	srvcList := ServicesList{}
	srvcList.Init()
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
	srvc1 := Service{VIP: "1.1.1.1", Proto: "tcp", Port: "22"}
	rlSrv1 := RealServer{RIP: "1.1.1.2", Port: "22"}
	rlSrv2 := RealServer{RIP: "1.1.1.3", Port: "22"}
	rlSrv3 := RealServer{RIP: "1.1.1.4", Port: "22"}
	rlSrv4 := RealServer{RIP: "1.1.1.4", Port: "22"}
	srvc1.Init()
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
