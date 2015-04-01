#!/usr/bin/env python

try:
    from urllib import request as urllib_request
except ImportError:
    import urllib2 as urllib_request
import json
import sys


def main():
        url = sys.argv[1]
        data = json.dumps({"cmnd":"get info","service":"192.168.1.1"})
        try:
            request = urllib_request.Request(url,data)
            reader = urllib_request.urlopen(request)
            response = reader.read()
            print(response)
        except urllib_request.HTTPError:
            print("wrong request's type")


if __name__ == "__main__":
    main()

    
