## Golang based keepalive server

### Goals of the project
The main goal of the project is to build keepalived-like server with easy way to decouple control plane
and data path configuration (keepalived is tightly coupled with linux'es ipvs) and easy way to add new features
(i think that it's much easier to add something in Go program, than into C-bassed).
Other goals:
 try to implement other control plane features in main server (for example bgp-signaling to upstream peer that we 
 have live service; i want to try to write my own go-based tiny bgp implementation instead of using external bgp
 signaling (thru exabgp/bird/quagga))
 
 
### Current state:
  * basic configuration parser (it's working, not so many checks, but at least we can read cfg files and move on;
  it's not that hard to add new derectives into the parser; so rewriting it (or make it more "beautiful") isn't my main
  priority
  
  * right now we have implemented ipvsadm adapter (adapter in our cases is a way to program the forwarding path;
  right now (i wasnt able to find proper generic netlink implementation in Go) we are using ipvsadm bin (thru os/exec)
  for seting up forwarding path) lots of feature must be added; right now we are using it just for POC.
  
  * tcp and http/https healtchecks
  
  * signaling that real server has failed it's check/or check was successful
  
basicly that gives us a minimal feature set to move on with other parts of the project.

### Future plans and TODOs:
  * add bgp speaker (work in progress; https://github.com/tehnerd/go_bgpcp)
  * add external api for services configuration (rest,grpc etc)
  * stability and features improvement (lots of things must be added)
  * proper IPv6 support (MUSTMUSTMUST; right now we dont check/parse etc addresses of the reals and vip's)
  * ...
