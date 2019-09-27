# perf2cloudprofiler

Uploads perf output to Google Cloud Profiler.

## Installation

Linux 64-bit:

```
$ curl http://storage.googleapis.com/jbd-releases/perf2cloudprofiler > perf2cloudprofiler && chmod +x perf2cloudprofiler
```

## Usage

```
$ perf record tree /usr # profile tree
$ perf report           # inspect the profile
Samples: 4K of event 'cpu-clock:uhH', Event count (approx.): 1236250000
Overhead  Command  Shared Object     Symbol
  12.52%  tree     tree              [.] 0x000000000000c964
  11.06%  tree     libc-2.27.so      [.] vfprintf
   4.63%  tree     libc-2.27.so      [.] write
   4.33%  tree     libc-2.27.so      [.] __lxstat64
   3.84%  tree     libc-2.27.so      [.] _IO_file_xsputn
   3.01%  tree     libc-2.27.so      [.] __fprintf_chk
   2.75%  tree     libc-2.27.so      [.] __strcoll_l
   2.43%  tree     libc-2.27.so      [.] c32rtomb
   1.92%  tree     libc-2.27.so      [.] malloc
   1.80%  tree     libc-2.27.so      [.] 0x000000000016af98
   1.72%  tree     libc-2.27.so      [.] 0x00000000000e00a8
   1.62%  tree     libc-2.27.so      [.] 0x000000000018e5a5
   1.54%  tree     libc-2.27.so      [.] cfree
   1.38%  tree     libc-2.27.so      [.] iswprint
   1.01%  tree     libc-2.27.so      [.] 0x000000000018e590
   0.87%  tree     libc-2.27.so      [.] __open_nocancel
   0.87%  tree     libc-2.27.so      [.] 0x0000000000169d90
   0.75%  tree     libc-2.27.so      [.] __close_nocancel
   0.75%  tree     libc-2.27.so      [.] __fxstat64
   0.75%  tree     libc-2.27.so      [.] readdir64
...

$ perf2cloudprofiler -target=tree-profile
https://console.cloud.google.com/profiler/tree-profile;type=CPU?project=PROJECT
```

![Cloud Profiler Screenshot](https://i.imgur.com/4jsjxzJ.png)

## Known issues

* Users need to install [perf_to_profile](https://github.com/google/perf_data_converter) in their PATH.
* perf to pprof converter drops the symbol names, see [google/perf_data_converter#81](https://github.com/google/perf_data_converter/issues/81).
