## Golang based keepalive server
[developer's documentation](https://github.com/tehnerd/go_keepalived/blob/master/DEVDOCS.md)
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
  
  * so far we have implemented two adapters (adapter in our case is a way to program forwarding path):
    ipvsadm (limited features supported; we are using it thru os/exec) and generic netlink msgs (thru github.com/tehnerd/gnl2go lib)
    by default we are using ipvsadm (check cfg example inside main for more info about how to use netlink instead of ipvsadm)

  * tcp and http/https healtchecks
  
  * signaling that real server has failed it's check/or check was successful

  * bgp speaker/injector. now we can advertise/withdraw v4 routes into routing domain
  (work in progress; https://github.com/tehnerd/bgp2go)
  * initial IPv6 support (v6 vips + v6reals; TODO: check that v4 vips/v6 reals works with netlink). we can advertise VIPs to bgp peers as well.
  * initial support for http api (todo: httpS; more features)
  * testing support for zookeper's based healthcheck (actually just fooling around with zookeeper; IMO this is wrong way to implement ZK's support.
    it will make more sense to make separate, ZK bassed tool, which will pull all the configuration (IPs of VIPs, state of reals, other type of 
    cfg from) from ZK, and wont have any cfg file at all (or just with few lines - IPs of ZK itself). If you are interesting in this type of tool - drop me a not
    and/or PR)
 
basicly that gives us a minimal feature set to move on with other parts of the project.

### Future plans and TODOs:
  * draft of howto/tutorial
  * add more features into http api
  * add other external apis for services configuration (grpc,thrift(if i able to find proper go's
     thrift's compiler etc)
  * stability and features improvement (lots of things must be added)
  * ...
