
# Simple requests using curl

```
$ curl http://127.0.0.1:8000/events\?clear=1
[]
$ curl -d"@event.json" http://127.0.0.1:8000
```

# Load testing with Apache ab

100 POST requests, max 10 concurrent, no use of keep-alive:

```
$ ab -n 100 -c 10 -p event.json http://127.0.0.1:8000/
This is ApacheBench, Version 2.3 <$Revision: 1826891 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking 127.0.0.1 (be patient).....done


Server Software:        
Server Hostname:        127.0.0.1
Server Port:            8000

Document Path:          /
Document Length:        0 bytes

Concurrency Level:      10
Time taken for tests:   0.009 seconds
Complete requests:      100
Failed requests:        0
Total transferred:      7500 bytes
Total body sent:        104600
HTML transferred:       0 bytes
Requests per second:    10896.81 [#/sec] (mean)
Time per request:       0.918 [ms] (mean)
Time per request:       0.092 [ms] (mean, across all concurrent requests)
Transfer rate:          798.11 [Kbytes/sec] received
                        11130.92 kb/s sent
                        11929.02 kb/s total

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.0      0       0
Processing:     0    1   0.3      1       1
Waiting:        0    1   0.3      1       1
Total:          0    1   0.2      1       2

Percentage of the requests served within a certain time (ms)
  50%      1
  66%      1
  75%      1
  80%      1
  90%      1
  95%      1
  98%      1
  99%      2
 100%      2 (longest request)
```

So, 0.92ms per request :)

If we add -k to use http keep-alive then the mean POST time drops to 0.041ms:

```
$ ab -k -n 10000 -c 10 -p event.json http://127.0.0.1:8000/
This is ApacheBench, Version 2.3 <$Revision: 1826891 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking 127.0.0.1 (be patient)
Completed 1000 requests
Completed 2000 requests
Completed 3000 requests
Completed 4000 requests
Completed 5000 requests
Completed 6000 requests
Completed 7000 requests
Completed 8000 requests
Completed 9000 requests
Completed 10000 requests
Finished 10000 requests


Server Software:        
Server Hostname:        127.0.0.1
Server Port:            8000

Document Path:          /
Document Length:        0 bytes

Concurrency Level:      10
Time taken for tests:   0.411 seconds
Complete requests:      10000
Failed requests:        0
Keep-Alive requests:    10000
Total transferred:      990000 bytes
Total body sent:        10700000
HTML transferred:       0 bytes
Requests per second:    24354.96 [#/sec] (mean)
Time per request:       0.411 [ms] (mean)
Time per request:       0.041 [ms] (mean, across all concurrent requests)
Transfer rate:          2354.63 [Kbytes/sec] received
                        25449.03 kb/s sent
                        27803.66 kb/s total

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.0      0       0
Processing:     0    0   0.7      0      10
Waiting:        0    0   0.7      0      10
Total:          0    0   0.7      0      10

Percentage of the requests served within a certain time (ms)
  50%      0
  66%      0
  75%      0
  80%      0
  90%      1
  95%      1
  98%      2
  99%      4
 100%     10 (longest request)
```

100 GET requests (with 100 events), max 10 concurrent:

```
$ ab -n 100 -c 10 http://127.0.0.1:8000/events
This is ApacheBench, Version 2.3 <$Revision: 1826891 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking 127.0.0.1 (be patient).....done


Server Software:        
Server Hostname:        127.0.0.1
Server Port:            8000

Document Path:          /events
Document Length:        89302 bytes

Concurrency Level:      10
Time taken for tests:   0.046 seconds
Complete requests:      100
Failed requests:        0
Total transferred:      8938300 bytes
HTML transferred:       8930200 bytes
Requests per second:    2192.65 [#/sec] (mean)
Time per request:       4.561 [ms] (mean)
Time per request:       0.456 [ms] (mean, across all concurrent requests)
Transfer rate:          191391.86 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.0      0       0
Processing:     2    4   0.8      4       6
Waiting:        2    4   0.7      4       6
Total:          3    4   0.8      4       6

Percentage of the requests served within a certain time (ms)
  50%      4
  66%      4
  75%      5
  80%      5
  90%      5
  95%      6
  98%      6
  99%      6
 100%      6 (longest request)
```

With 1100 events, GET time reduces from 0.45ms to 2.4653 ms per request.

```
Requests per second:    405.63 [#/sec] (mean)
Time per request:       24.653 [ms] (mean)
Time per request:       2.465 [ms] (mean, across all concurrent requests)
Transfer rate:          389147.84 [Kbytes/sec] received
```

However, POST time remains as above.
