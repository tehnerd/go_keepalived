import sys
import os
gen_py_path = '/tmp/'
thrift_pylib = '/tmp/'
if 'GEN_PY' in os.environ:
    gen_py_path = os.environ['GEN_PY']
if 'THRIFT_PYLIB' in os.environ:
    thrift_pylib = os.environ['THRIFT_PYLIB']
sys.path.append(gen_py_path)
sys.path.append(thrift_pylib)


#TODO: proper path to cfg_slb.py
from cfg_slb import slbAPI

from gokeepalived import Gokeepalived
from gokeepalived.ttypes import *



from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol
from thrift.server import TServer

class GokeepalivedHandler:
    def __init__(self):
        # 1 - link to http api; 2 - master pwd
        self._api = slbAPI(sys.argv[1],sys.argv[2])
        self.log = {}
    
    def api_call(self, request):
        response = self._api.ExecCmnd(request)
        return response["Data"]

    



handler = GokeepalivedHandler()
processor = Gokeepalived.Processor(handler)
transport = TSocket.TServerSocket(port=9090)
tfactory = TTransport.TBufferedTransportFactory()
pfactory = TBinaryProtocol.TBinaryProtocolFactory()

#server = TServer.TSimpleServer(processor, transport, tfactory, pfactory)

# You could do one of these for a multithreaded server
server = TServer.TThreadedServer(processor, transport, tfactory, pfactory)
#server = TServer.TThreadPoolServer(processor, transport, tfactory, pfactory)

print 'Starting the server...'
server.serve()
print 'done.'
