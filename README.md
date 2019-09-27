# perf2cloudprofiler

Uploads perf output to Google Cloud Profiler.

## Installation

Linux 64-bit:

```
$ curl http://storage.googleapis.com/jbd-releases/perf2cloudprofiler > perf2cloudprofiler && chmod +x perf2cloudprofiler
```

## Usage

```
$ perf record ls # profile ls
$ perf report
Samples: 2  of event 'cpu-clock:uhH', Event count (approx.): 500000
Overhead  Command  Shared Object     Symbol
  50.00%  ls       ld-2.27.so        [.] 0x000000000000b0ca
  50.00%  ls       libc-2.27.so      [.] 0x00000000000bbcac

$ perf2cloudprofiler -target=ls-profile
https://console.cloud.google.com/profiler/ls-profile;type=CPU?project=PROJECTID
```

![Cloud Profiler Screenshot](https://i.imgur.com/4jsjxzJ.png)
