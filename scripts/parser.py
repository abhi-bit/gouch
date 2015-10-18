#!/usr/bin/python

from matplotlib import pyplot as plt
import os
import re

tc = re.compile(r'\s*(\w+)\s*threads\s*\w+\s*(\w+)\sconnections*')
latency = re.compile(r'\s*Latency\s*(\w+.\w+)\s*(\w+.\w+)\s*(\w+.\w+)*')
req = re.compile(r'Requests/sec:\s+(\w+.\w+)*')
transfer = re.compile(r'Transfer/sec:\s+(\w+.\w+)*')

global_stats = dict()
go_limits = list()
go_latency_avg = list()
go_req_per_sec = list()
go_transfer_per_sec = list()

erl_limits = list()
erl_latency_avg = list()
erl_req_per_sec = list()
erl_transfer_per_sec = list()

def do_stats_gathering(stats_name, lang):

    rv = list()
    for filename in global_stats.keys():
        for stat in global_stats[filename].keys():
            if (stat == stats_name and
            global_stats[filename]["language"] == lang):
                rv.append(global_stats[filename][stats_name])
    return rv

def collate_results():

    for filename in os.listdir("."):
        if filename.endswith("log"):
            fd = open(filename)
            data = fd.read()
            t, c = re.findall(tc, data)[0]
            l_avg, l_stddev, l_max = re.findall(latency, data)[0]
            req_per_sec = re.findall(req, data)[0]
            transfer_sec = re.findall(transfer, data)[0]

            stat = dict()
            stat["limit"] = int(filename.split("-")[0])
            stat["language"] = filename.split("-")[1].split(".")[0]
            stat["threads"] = int(t)
            stat["connections"] = int(c)

            if "ms" in l_avg:
                stat["latency_avg"] = float(l_avg[:-2])
            elif "us" in l_avg:
                stat["latency_avg"] = 10000 #10s
            elif "s" in l_avg:
                stat["latency_avg"] = float(l_avg[:-1]) * 1000

            stat["latency_stddev"] = l_stddev
            stat["latency_max"] = l_max
            stat["request_per_sec"] = req_per_sec

            if "MB" in transfer_sec:
                stat["transfer_sec"] = float(transfer_sec[:-2])
            elif "KB" in transfer_sec:
                stat["transfer_sec"] = float(transfer_sec[:-2])/1000
            global_stats[filename] = stat

    #print json.dumps(global_stats)

def main():

    collate_results()

    go_limits = do_stats_gathering("limit", "go")
    go_transfer_per_sec = do_stats_gathering("transfer_sec", "go")
    go_latency_avg = do_stats_gathering("latency_avg", "go")
    go_req_per_sec = do_stats_gathering("request_per_sec", "go")

    erl_limits = do_stats_gathering("limit", "erl")
    erl_transfer_per_sec = do_stats_gathering("transfer_sec", "erl")
    erl_latency_avg = do_stats_gathering("latency_avg", "erl")
    erl_req_per_sec = do_stats_gathering("request_per_sec", "erl")

    print go_limits
    print go_transfer_per_sec
    print go_latency_avg
    print go_req_per_sec

    print erl_limits
    print erl_transfer_per_sec
    print erl_latency_avg
    print erl_req_per_sec

    plt.plot(go_limits, go_req_per_sec, label="Requests per sec")
    plt.plot(go_limits, go_latency_avg, label="latency in ms")
    plt.xlabel('limit')
    plt.legend()
    plt.show()

if __name__ == "__main__":
    main()
