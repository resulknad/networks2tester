package main

import "github.com/resulknad/networks2tester/test"
import "log"
import "os"
import "fmt"
import "flag"
import "time"

type TestInfo struct {
	Desc string
	N int
	Timeout int
	launch func(int,int) bool
	MaxRetry int
}

func (ti *TestInfo) launchTest() (err bool,success bool){
	defer func() {
		if r:= recover(); r!=nil {
			log.Println("recovered from ",r)
			err = true
			success = false
		}
	}()
	
	err = false
	success = ti.launch(ti.N,ti.Timeout)

	return
}

func (ti *TestInfo) Execute() (err bool, succ bool) {
	for i:=0; i<ti.MaxRetry; i++ {
		err,succ = ti.launchTest()
		if !err {
			return
		}
	}
	return
}


func main() {
	singleTest := flag.Int("t", -1, "run a single test")
	flag.Parse()
	
	f, err := os.OpenFile("tester.log", os.O_RDWR | os.O_CREATE | os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	
	tests := []TestInfo{
		TestInfo{Desc: "Fully connected, remove and add all", N:3, Timeout:60, launch:FullyConnectedDropSlowly, MaxRetry:2},
		TestInfo{Desc: "TwoCon", N:3, Timeout:60, launch:TwoconTopo, MaxRetry:2},
		TestInfo{Desc: "Small graph weight adjustment", N:0, Timeout:60, launch:BasicWeightAdjustment, MaxRetry:2},
		TestInfo{Desc: "complex.topo", N:0, Timeout:60, launch:ComplexTopo, MaxRetry:2},
		TestInfo{Desc: "star.topo", N:0, Timeout:60, launch:StarTopo, MaxRetry:2},
		TestInfo{Desc: "simple.topo", N:0, Timeout:60, launch:SimpleTopo, MaxRetry:2},
		TestInfo{Desc: "tri.topo", N:0, Timeout:60, launch:TriTopo, MaxRetry:2},
		TestInfo{Desc: "complex2.topo", N:0, Timeout:60, launch:Complex2Topo, MaxRetry:2},
		TestInfo{Desc: "Memory Leaks (only for interface changes)", N:3, Timeout:60, launch:MemUsage, MaxRetry:2},
		TestInfo{Desc: "Triggered Update (set down)", N:3, Timeout:60, launch:TriggeredUpdateRemove, MaxRetry:2},
		TestInfo{Desc: "Triggered Update (cost change)", N:3, Timeout:60, launch:TriggeredUpdateCost, MaxRetry:2},
		TestInfo{Desc: "Longest Prefix Matching", N:3, Timeout:60, launch:LongestPrefixMatching, MaxRetry:2}}
	
	runTest := func(i int,ti TestInfo) {
		log.Print("Starting Test " + ti.Desc)
		fmt.Printf("Test %d (%s): ",i,ti.Desc)
		err, suc := ti.Execute()
		if err {
			fmt.Println("(ERR)")
		} else {
			fmt.Printf("%t\n", suc)
		}
		time.Sleep(1*time.Second)
	}

	if *singleTest != -1 {
		runTest(*singleTest, tests[*singleTest])	
		return
	}
	for i,ti := range tests {
		runTest(i,ti)
	}

}

func TwoconTopo(n int, timeout int) bool {
	t := ParseTopoFile("twocon.topo")
	defer t.TearDown()

	t.DrawGraph("out.svg")
	t.StartTest()

	if !t.WaitUntilCorrect(timeout) {
		return false
	}
	return true
}
func ComplexTopo(n int, timeout int) bool {
	t := ParseTopoFile("complex.topo")
	defer t.TearDown()

	t.DrawGraph("out.svg")
	t.StartTest()

	if !t.WaitUntilCorrect(timeout) {
		return false
	}
	return true
}

func Complex2Topo(n int, timeout int) bool {
	t := ParseTopoFile("complex2.topo")
	defer t.TearDown()

	t.DrawGraph("out.svg")
	t.StartTest()

	if !t.WaitUntilCorrect(timeout) {
		return false
	}
	return true
}

func SimpleTopo(n int, timeout int) bool {
	t := ParseTopoFile("simple.topo")
	defer t.TearDown()

	t.DrawGraph("out.svg")
	t.StartTest()

	if !t.WaitUntilCorrect(timeout) {
		return false
	}
	return true
}

func StarTopo(n int, timeout int) bool {
	t := ParseTopoFile("star.topo")
	defer t.TearDown()

	t.DrawGraph("out.svg")
	t.StartTest()

	if !t.WaitUntilCorrect(timeout) {
		return false
	}
	return true
}

func TriTopo(n int, timeout int) bool {
	t := ParseTopoFile("tri.topo")
	defer t.TearDown()

	t.DrawGraph("out.svg")
	t.StartTest()

	if !t.WaitUntilCorrect(timeout) {
		return false
	}
	return true
}

func TriggeredUpdateRemove(N int, timeout int) bool {
  return TriggeredUpdate(N, timeout, true)
}
func TriggeredUpdateCost(N int, timeout int) bool {
  return TriggeredUpdate(N, timeout, false)
}
func TriggeredUpdate(N int, timeout int, remove bool) bool {
	t := test.NewTest()
	defer t.TearDown()
	routers := []*test.Router{}
	for i:=0; i<N; i++ {
	  routers = append(routers,t.AddRouter())
	}

	// this adds an interface to d1, which isn't connected to any router...
	ip,mask := t.NextAddressSpace()
	subnet := t.GetOrCreateSubnet(ip,mask)
	intrf := t.GetOrCreateInterface(routers[0], ip, mask, subnet)
	t.SetInterfaceCost(intrf, 1)

	// creates a triangle, but bc of cost 16, traffic can only move in one direction
	// advertisments in both directions
	for i:=0; i<N; i++ {
		ip2,mask2 := t.NextAddressSpace()
		t.ConnectRoutersUni(routers[((i-1)+N)%(N)], routers[i], 1, ip2, mask2, ip2+1)
		t.ConnectRoutersUni(routers[i],routers[((i-1)+N)%(N)],  16, ip2+1, mask2, ip2)
	}	 
	
	t.DrawGraph("out.svg")
	t.StartTest()

	// wait for initial convergence
	if !t.WaitUntilCorrect(timeout) {
		return false
	}

	if remove {
	  t.TakeDownInterface(intrf) 
	} else {
	  t.SetCostGraphAndInstance(intrf, 17)
	}

	// if triggered updates arent implemented,
	// this will converge very slowly now
	// for route to subnet 1.1.2.0 the next hops are (w/o triggered updates):
	// d1: d3
	// d2: d1
	// d3: d2
	// with triggered updates, this doesnt happen.

	t.DrawGraph("out.svg")
	if !t.WaitUntilCorrect(5) {
		return false
	}
	return true
}

func LongestPrefixMatching(N int, timeout int) bool {
	
	lengthToMask := func(n uint) uint32 { return (0xFFFFFFFF)<<(32-n) }
	ip := test.Ip2int("123.123.123.123")

	t := test.NewTest()
	defer t.TearDown()

	router := t.AddRouter()
	
	var i uint
	for i=8; i<30; i+=5 {
	  rt := t.AddRouter()
	  t.ConnectRouters(router,rt,1,1)

	  mask := lengthToMask(i)
	  subnet := t.GetOrCreateSubnet(ip,mask)
	  intrf := t.GetOrCreateInterface(rt, ip, mask, subnet)
	  t.SetInterfaceCost(intrf, float64(i))
	  ip+=1
	}	 
	
	t.DrawGraph("out.svg")
	t.StartTest()

	return t.WaitUntilCorrect(timeout)
}

func BasicWeightAdjustment(n int, timeout int) bool {
	t := test.NewTest()
	defer t.TearDown()
	routers := []*test.Router{}
	for i:=0; i<3; i++ {
		routers = append(routers,t.AddRouter())
	}
	for i:=0; i<3; i++ {
		for j:=i+1; j<3; j++ {
			if i!=j {
				t.ConnectRouters(routers[i], routers[j], 1, 1)
			}
		}
	}
	
	t.DrawGraph("out.svg")
	t.StartTest()

	if !t.WaitUntilCorrect(timeout) {
		return false
	}
	
	var c float64
	for c=12; c<17; c++ {
		for i:=0; i<3; i++ {
			for j:=0; j<3; j++ {
				if i!=j {
					t.SetCostUni(routers[i], routers[j], c)
					t.DrawGraph("out.svg")
					if !t.WaitUntilCorrect(timeout) {
						return false
					}
				}
			}
		}
	}

	return true
}

func MemUsage(n int, timeout int) bool {
	t := test.NewTest()
	defer t.TearDown()
	routers := []*test.Router{}

	for i:=0; i<n; i++ {
		routers = append(routers,t.AddRouter())
	}

	for i:=0; i<n; i++ {
		for j:=i+1; j<n; j++ {
			if i!=j {
				t.ConnectRouters(routers[i], routers[j], 1, 1)
			}
		}
	}

	t.DrawGraph("out.svg")
	t.StartTest()
	initialMemUsage := t.MemUsage()

	for k:=0; k<20000; k++ {
	  for i:=0; i<n; i++ {
		  for j:=0; j<n; j++ {
			  if i!=j {
				  t.TakeDownUni(routers[i], routers[j])
			  }
		  }
	  }

	  for i:=0; i<n; i++ {
		  for j:=0; j<n; j++ {
			  if i!=j {
				  t.PutUpUni(routers[i], routers[j])
			  }
		  }
	  }
	}
	
	maxChange := 0.
	finalMemUsage := t.MemUsage()
	for i, m := range initialMemUsage {
	  diff := (finalMemUsage[i] - m)
	  diffP := (diff/m)*100
	  log.Printf("Memory changed by %f (%f%%)",diff, diffP)
	  if diffP > maxChange {
		maxChange = diffP
	  }
	}
	return maxChange < 120
  }

func FullyConnectedDropSlowly(n int, timeout int) bool {
	t := test.NewTest()
	defer t.TearDown()
	routers := []*test.Router{}
	for i:=0; i<n; i++ {
		routers = append(routers,t.AddRouter())
	}
	for i:=0; i<n; i++ {
		for j:=i+1; j<n; j++ {
			if i!=j {
				t.ConnectRouters(routers[i], routers[j], 1, 1)
			}
		}
	}

	t.DrawGraph("out.svg")
	t.StartTest()
	if !t.WaitUntilCorrect(timeout) {
		return false
	}

	for i:=0; i<n; i++ {
		for j:=0; j<n; j++ {
			if i!=j {
				t.TakeDownUni(routers[i], routers[j])
				t.DrawGraph("out.svg")
				if !t.WaitUntilCorrect(timeout) {
					return false
				}
			}
		}
	}

	for i:=0; i<n; i++ {
		for j:=0; j<n; j++ {
			if i!=j {
				t.PutUpUni(routers[i], routers[j])
				t.DrawGraph("out.svg")
				if !t.WaitUntilCorrect(timeout) {
					return false
				}
			}
		}
	}
	return true
}
