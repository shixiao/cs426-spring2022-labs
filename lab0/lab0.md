# [Initial version] CS426 Lab 0: Introduction to Go and multi-threaded programming

## Overview
The goal of this lab is to familiarize you with the Go programming language, multi-threaded programming, and the concepts and practices of concurrency and parallelization. Have fun!

## Logistics
**Policies**
- Lab 0 is meant to be an **individual** assignment. Please see the [Collaboration Policy](../collaboration_policy.md) for details.
- We will help you strategize how to debug but WE WILL NOT DEBUG YOUR CODE FOR YOU.
- Please keep and submit a time log of time spent and major challenges you've encountered. This may be familiar to you if you've taken CS323. See [Time logging](../time_logging.md) for details.

- Questions? post to Canvas or email the teaching staff at cs426ta@cs.yale.edu.
  - Richard Yang (yry@cs.yale.edu)
  - Xiao Shi (xiao.shi@aya.yale.edu)
  - Scott Pruett (spruett345@gmail.com) (unavailable until Jan 30th)

**Submission deadline: 23:59 ET Wednesday Feb 9, 2022**

**Submission logistics** TBA.

Your submission for this lab should include the following files:
```
discussions.txt
profile_striped_num_cpu_times_two.png
string_set.go
striped_string_set.go
time.log
```

## Preparation
1. Install Go (or run Go on a Zoo machine) following [this guide]().
2. Complete the [Tour of Go](https://go.dev/tour/list). Pay special attention to the concurrency module. Use the exercises in the tutorial as checkpoints of your Go familiarity, but we won't as you to submit your solutions or grade them. That said, skip them at your own peril as all of the labs this semster will be in Go.
3. Check out the list of [Go tips and FAQs](#go-tips-and-faqs) as you set up your development workflow.

## Part A. Thread-safe add-only string set
Build a data structure `LockedStringSet` which implements the `StringSet` interface (in `string_set.go`). Use `sync.RWMutex` to ensure that the data structure is thread-safe.

You can now run the provided simple unittest and possibly add your own. `-v` turns on the verbose flag which provides more information about the test run. `-race` turns on the [Go race detector](https://go.dev/doc/articles/race_detector), which is a tool that detects data races.
```
go test string_set -race -v
```

## Part B. Lock striping

**Benchmarks** and **Profilers** are important tools to evaluate performance of one's code, protocol, or systems. Single-machine, small-scale benchmarks are often called _micro-benchmarks_. Profilers collect the runtime metrics of a program (often a benchmark) to guide performance optimizations. The most common profiles include CPU profiles and memory profiles. Benchmarking and profiling can be somewhat of a minefield since it's easy to be measuring the wrong thing or drawing incorrect conclusions, e.g., sometimes one's benchmark setup take more CPU cycles to run than the core logic of interest. We will only get a taste of it in this lab.

The following command runs the benchmark with names matching `Locked` (`-bench=Locked`), prevents tests from running (`-run=^$`), collects the CPU profile, and outputs the result into `profile.out`.
```
go test -v -bench=Locked -run=^$ -benchmem -cpuprofile profile.out
```
To view the profile, run `go tool pprof profile.out`. type `top` to view the most time consuming functions, and `png` to generate a graph (you need to install [graphviz](https://graphviz.org/download/)). The lab0 repo includes one such example graph.
```
$  go tool pprof profile.out

Type: cpu
Time: ... (EST)
Duration: 1.50s, Total samples = 4.42s (294.13%)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top
Showing nodes accounting for 3800ms, 85.97% of 4420ms total
Dropped 25 nodes (cum <= 22.10ms)
Showing top 10 nodes out of 58
      flat  flat%   sum%        cum   cum%
    2290ms 51.81% 51.81%     2290ms 51.81%  runtime.usleep
     740ms 16.74% 68.55%     1560ms 35.29%  sync.(*Mutex).lockSlow
     160ms  3.62% 72.17%      160ms  3.62%  runtime.pthread_cond_wait
     150ms  3.39% 75.57%      240ms  5.43%  runtime.mapaccess2_faststr
     150ms  3.39% 78.96%     1400ms 31.67%  sync.(*RWMutex).Unlock
      80ms  1.81% 80.77%     1720ms 38.91%  runtime.lock2
      60ms  1.36% 82.13%       60ms  1.36%  runtime.asmcgocall
      60ms  1.36% 83.48%       60ms  1.36%  runtime.cansemacquire (inline)
      60ms  1.36% 84.84%       60ms  1.36%  runtime.pthread_cond_signal
      50ms  1.13% 85.97%       50ms  1.13%  runtime.tophash (inline)
(pprof) png
Generating report in profile001.png
```

As you can see in the sample output (`profile001.png`), the single RWMutex has become the bottleneck on the data structure (`runtime.usleep` is part of the RWMutex locking and waiting strategy to aoid thrashing or busy wait). "Lock striping" is a common technique to get around lock contention of a coarse-grained lock.

The idea is to split the data structure into "stripes" (or "buckets", or "shards"), each protected by its own finer-grained lock. Many queries can then be either parallelized or served from a single stripe. Typically, stripes are divided based on a good hash of the key or id (in our case, the string).

**B1.** Implement `StripedStringSet` which offers the same API as `StringSet` but uses lock striping. Add a factory function that constructs a `StripedStringSet` with a given stripe count. (Starter code in `striped_string_set.go`.)

You may find it helpful to review the following references to understand the technique, but you may not directly use or translate the code:
* [Lock striping in Java](https://www.baeldung.com/java-lock-stripping)
* [Concurrent map in Go](https://github.com/orcaman/concurrent-map/blob/master/concurrent_map.go)

**B2.** Keeping count. There are several strategies to keep count in this striped data structure:
1. iterate through the entire set across all stripes each time a count is needed;
2. keep a per-stripe counter and add up counters from every stripe when queried;
3. keep a global (i.e., one single) counter that gets updated when a string is added; (See [atomic counters](https://gobyexample.com/atomic-counters).)

What are the advantages and disadvantages of each of these approaches? Can you think of query patterns where one works better than another?

Include your thoughts (1~2 paragraphs) in a plain text file `discussions.txt` under a heading `B2`.

**ExtraCredit1.** Suppose we start with a `StripedStringSet` with x unique strings. Goroutine/thread 0 issues a `Count()` call, while threads 1~N issues `Add()` calls with distinct strings. What values might the `Count()` in thread 0 return? Why? Does it matter which counting strategy (#1~3 above) we use? What about `LockedStringSet`? In light of this behavior, how might you define "correctness" for the method `Count()`? Include your thoughts in `discussions.txt` under a heading `ExtraCredit1`.

## Part C. Channels, goroutines, and parallelization
**C1.** Implement the API `PredRange(begin, end, pattern)` to return all strings matching a particular pattern within a range `[begin, end)` lexicographically.

For example, if the string set `s` contains `{"fooabc", "barabc", "bazdef", "tusabc", "zyxabc"}`, calling `s.PredRange("barabc", "zyxabc", "abc")` should return `["barabc", "tusabc"]`.

You may use [`regexp.Match`](https://pkg.go.dev/regexp).

**C2.** Parallelize your implementation of `PredRange` by spinning up a goroutine for each stripe and aggregate the results in the end.

**C3.** Run the provided benchmark now and see the difference in performance between `LockedStringSet` and `StripedStringSet` with 2 stripes (`BenchmarkStripedStringSet2`). What do you observe? Include the results in `discussions.txt` under a heading `C3`.

**C4.** Generate a graph visualization of a profile for `StripedStringSet` with `stripeCount == NumCPU() * 2`. Name this `profile_striped_num_cpu_times_two.png`.

**ExtraCredit2.** Discuss the effect of the parameter stripeCount on the performance (compared to `LockedStringSet`). What do you notice? Why? What's the optimal stripeCount (feel free to try other numbers and include the result in the discussion)? Include your thoughts in `discussions.txt` under a heading `ExtraCredit2`.

# End of Lab 0
---

# Go Tips and FAQs
 - After the Tour of Go, use https://go.dev/doc/effective_go and https://gobyexample.com/
 - Use these commands:
    - `go fmt`
    - `go mod tidy` cleans up the module dependencies.
    - `go test -race ...` turns on the Go race detector.
    - `go vet` examines Go source code and reports suspicious constructs, such as Printf calls whose arguments do not align with the format string. Vet uses heuristics that do not guarantee all reports are genuine problems, but it can find errors not caught by the compilers.
