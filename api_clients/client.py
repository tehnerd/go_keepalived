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

from gokeepalived import Gokeepalived
from gokeepalived.ttypes import *

from thrift import Thrift
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol

try:
  # Make socket
    transport = TSocket.TSocket('localhost', 9090)

  # Buffering is critical. Raw sockets are very slow
    transport = TTransport.TBufferedTransport(transport)

  # Wrap in a protocol
    protocol = TBinaryProtocol.TBinaryProtocol(transport)

  # Create a client to use the protocol encoder
    client = Gokeepalived.Client(protocol)

  # Connect!
    transport.open()
    data = {"Command":"GetInfo"}
    response = client.api_call(data)
    print(response)
    transport.close()

except Thrift.TException, tx:
  print '%s' % (tx.message)
