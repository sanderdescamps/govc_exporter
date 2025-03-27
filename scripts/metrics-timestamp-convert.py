#!/usr/bin/env python

import sys
import time
import re


def convert_timestamps(metrics_text:str):
    for line in metrics_text.splitlines():
        if line.startswith("#"):
            continue
        pattern = re.compile(r'^(\w+\{.*\}) ([\d\.e\+\-]+) ([\d]+)$', re.MULTILINE)
        m = pattern.match(line)
        timestamp = time.gmtime(int(m.group(3))/1000)
        lineNiceTime = "%s %s %s"%(m.group(1), m.group(2), time.strftime('%Y-%m-%d %H:%M:%S', timestamp) )
        yield lineNiceTime

def main():
    data = sys.stdin.read()
    for line in convert_timestamps(data):
        print(line)

if __name__ == "__main__":
    main()
