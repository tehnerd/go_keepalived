###### Testing on local node w/o bgp and ipvsadm support:
for testing purpose you can add "testing" as a first line in cfg file. this stanza will instruct adapter module to use dummy adapter instead
of ipvsadm. also you need to change notifier's name from bgp to something else (in that case we will start dummy notifier instead of bgp2go)

##Project's architecture
![alt text](https://github.com/tehnerd/go_keepalived/blob/master/structure.jpg?raw=true "Project's architecture")

#### Startup process
During startup we are reading configure from the cfg file (example could be found @ main/gokeepalived_test.cfg) with
cfgparser (cfgparser/cfgparser.go) using ReadCfg routine.

This routine is responsible for the:
 1. parse config and exit the program if something goes wrong
 2. populate service.ServicesList, notifier.NotifierConfig and api.GenericAPI structs (more about this structs bellow)
 3. Starts API's goroutine at the end of the ReadCfg

after these steps control goes back to the main function.

main routine is going to start ServicesList goroutine and will wait indefinitely

#### Modules description
High lvl overview of program's architecture could be found on the picture above.

Description of each module (as well as it's configuration parts etc could be found in sections bellow)


#####Api module (api/api.go)
Api module is responsible for the communication with the outside world. For example we could add/remove
services and reals with api cals from the external node.

Depending on cfg file we could have any type of api endpoints. Right now only http is implemented (api/http_api.go).
New endpoints could be easily added. The must communication with Api thru APIMsg chans.
```go
type APIMsg struct {
    Cmnd string //not used right now, instead we are using Data["Command"]; prob will be removed
    Data *map[string]string
}
```
API's endpoint must convert external request to struct map[string]string and must insure that there is exists "Command" 
key in this struct.


API module will parse the request (ProcessApiRequest routine): make sanity check (that all requered fields for particular 
"Command" are there, if during the startup master key was provided, it's going to check the msg's digest.

at any point, if there was error during the parsing, it will return the result to endpoint for further processing.

if there wasnt any error, the parsed request is going to be send to the ToServiceList chan of type service.ServiceMsg
```go
type ServiceMsg struct {
    Cmnd    string
    Data    string
    DataMap *map[string]string
}
```

after that API will wait for the response from the FromServiceList chan and send it back to the endpoint.

#####ServicesList module (service/service.go and service/msgs2service_processing.go; struct ServicesList)
After startup this module is responsible for the processing of msgs from the API. (thru the DispatchSLMsgs routine in 
msgs2service_processing.go).

On receiving of msg from api it's going to parse it and send it further  (according to the command inside the msg)
to either goroutine, responsible for the particular (described in msg) service, or notifier (if, for example we would like 
to stop advertise availability of the service to the world)

In case of service, msg is going to be send to the ToService chan (type ServiceMsg) and in case of notifier - to ToNotifier chan (type notfier.NotifierMsg)
```go
type NotifierMsg struct {
    Type string
    Data string
}
```

#####Service module (service.go; Service struct)
Service module is responsible for the bookkeeping of the service's related information
(such as: how many reals alive; if service has enough alive reals; does it exceed quorum etc; for more info check Service struct)
the code is pretty much selfexplained. we run one goroutine for each of the service (check func StartService() in service.go)
and listens for events from:
 1. ServicesList: thru ToService chan. such as add/remove real. or shutdown the service.
 2. RealServers: thru FromReal chan. such as alive/dead or fatal error in real

and notify adapter about this events (to add/remove alive/daed reals in forwarding path; or to send msg thru notifier)


#####RealServer module (service.go; RealServer struct)
RealServer module is responsible for bookkeeping information about healthcheks results toward this real
it listens for:
 1. Events from healthcheck (if it was successfull or not).and if server became alive -> send this info to the Service goroutine.
 2. Events from Service (right now there is only one event: to shutdown/remove real from the pool. in future we could change checks and/or weight of the real etc)

During startup this module starts healthcheck goroutine of particular (according to the configuration) type.
right now we support tcp healthcheck and http/https (healtchecks/tcp_check.go and http_check.go)


#####Adapter module (adapter/adapter.go adapter/ipvsadm_adapter.go)
Adapter is responsible for the programming of forwarding path.

right now it also acts as a proxy to notifier module

currently we supports  ipvsadm (and therefore linux's lvs; adapter/ipvsadm_adapter.go) and dummy (which just prints recved msgs. using it for the testing)
as a way to programm the forwarding path.
but it shouldn't be hard to rewrite this part if something else will be needed.

Adapter listens for msgs on chan type AdapterMsg.

Received msgs will be parsed with IPVSAdmExec routine and the will be proxied to the notifier (right now we dont destinct if msg is destined  to adapter or notifier;
if either of em doesnt support msg's type it just ignores it. for example adapter ignores "AdvertiseService" msg's type and notifier ignores "AddReal")


#####Notifier module (notifier/notifier.go notifier/bgp_notifier.go)
This module is responsible for notification of "outside world" about health of Service on local node. Right now we are using bgp for that purpose
(thru github.com/tehnerd/bgp2go lib) (and dummy notifier if case of testing; which just prints all rcved msgs)

this module is listening for the msgs on chan NotifierMsg, parse em, and send parsed commands to bgp2go lib.
for example we could either advertise host-route to local bgp peer, if service became "alive" or "withdraw" it, if there is zero (or less than quorum) 
alive realservers.

all suported commands could be found in notifier/bgp_notifier.go and in future will be described in user's guide.


#####Misc:
example of http's api client could be found @ api_clients/cfg_slb.py (TODO: add comments with example of all supported commands.
i'm afraid right now you need to check api/api.go for supported commands)


