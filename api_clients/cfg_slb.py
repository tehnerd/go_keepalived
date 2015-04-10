#!/usr/bin/env python

try:
    from urllib import request as urllib_request
except ImportError:
    import urllib2 as urllib_request
import json
import sys
import hmac
import hashlib

class slbAPI(object):

    def __init__(self, endpoint, password):
        self._endpoint = endpoint
        self._password = password

    def ExecCmnd(self,cmnd_data):
        line = str()
        keyList = list()
        for key in cmnd_data:
            keyList.append(key)
        for key in sorted(keyList):
            line = "".join((line,cmnd_data[key]))
        cmnd_data["Digest"] = hmac.new(self._password, line, digestmod=hashlib.sha256).hexdigest()
        data = json.dumps(cmnd_data)
        try:
            request = urllib_request.Request(self._endpoint,data)
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
        '''
        examples of supported commands(lots of copy paste right now while i'm writing new features
        and testing the output; gona remove it later)
        '''
        url = sys.argv[1]
        slb = slbAPI(url,"123")
        data = {"Command":"GetInfo","service":"192.168.1.1"}
        resp = slb.ExecCmnd(data)
        print(resp)
        data = {"Command":"AddService","VIP":"[fc12:1::2]","Port":"8080","Proto":"tcp"}
        resp = slb.ExecCmnd(data)
        print(resp)
        data = {"Command":"GetInfo","service":"192.168.1.1"}
        resp = slb.ExecCmnd(data)
        print(resp)
        data = {"Command":"StartAllNotification","VIP":"192.168.1.1"}
        resp = slb.ExecCmnd(data)
        print(resp)










if __name__ == "__main__":
    main()

    
