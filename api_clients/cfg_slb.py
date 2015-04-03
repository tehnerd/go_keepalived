#!/usr/bin/env python

try:
    from urllib import request as urllib_request
except ImportError:
    import urllib2 as urllib_request
import json
import sys

class slbAPI(object):

    def __init__(self, endpoint):
        self._endpoint = endpoint

    def ExecCmnd(self,cmnd_data):
        try:
            request = urllib_request.Request(self._endpoint,cmnd_data)
            reader = urllib_request.urlopen(request)
            api_response = reader.read()
            resp = json.loads(api_response)
            return resp
        except urllib_request.HTTPError:
            #TODO: more meaningfull comment; more errors handling
            return {"result":"error during api urllib_request"}
        except:
            return {"result":"generic error"}

        


def main():
        url = sys.argv[1]
        slb = slbAPI(url)
        data = json.dumps({"Command":"GetInfo","service":"192.168.1.1"})
        resp = slb.ExecCmnd(data)
        print(resp)


if __name__ == "__main__":
    main()

    
