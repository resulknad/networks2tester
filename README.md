networks2tester
===============
**Do not expect this tester to be 100% correct.**

Unofficial tester for second networks project. It constructs a graph, runs dijkstra on it and checks your implementation (whether all your next_hops are correct). Currently it doesn't look at your routing table. Supports dynamically adjusting weights and dropping links.

Tester still needs testing.

## Installation
### Binary
https://github.com/resulknad/networks2tester/releases

### From Source
Install golang 1.10 (older versions might work as well, didn't test) and set path variables as described in installation instructions of golang.

For the graph visualization the graphviz package is required. Tester might fail if dot executable isn't available on your system.

```bash
go get -u github.com/resulknad/networks2tester
```

## Usage
Open a terminal and go into your folder where libdr.so, dr, libvns and all the topo files reside.

For all test cases execute:
```bash
networks2tester
```

For a single test case:
```bash
networks2tester -t 1
```

Feedback on stdout:
```bash
Test 0 (Fully connected, remove and add all): true
Test 1 (Small graph weight adjustment): true
Test 2 (complex.topo): true
Test 3 (star.topo): true
Test 4 (simple.topo): true
Test 5 (tri.topo): true
Test 6 (complex2.topo): true
```
true means passed, false not passed and (ERR) is either an error in your dr_api or the tester failed.

For debugging purposes you may want to look at:
- out.svg which contains a visualization of the network graph
- tester.log where you can find information about why you didn't pass a test or why the tester crashed
- {ROUTERNAME}.txt files where the tester stores whatever you print

## Additional Testcases
If you want to add a testcase, take a look at main.go. There are two options:
### .topo file
The tester is able to parse the .topo files, (atm only link add and node add). This will allow for convergence tests, without dynamically removing edges or adjusting weights.
### Test API
In main.go you find two examples of how to use the Test API to construct your own tests. See FullyConnectedDropSlowly and BasicWeightAdjustment.
