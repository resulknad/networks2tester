package main

import "github.com/resulknad/networks2tester/test"
import "log"
import "os"
import "fmt"
import "flag"

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
		TestInfo{Desc: "Fully connected, remove and add all", N:4, Timeout:60, launch:FullyConnectedDropSlowly, MaxRetry:2},
		TestInfo{Desc: "Small graph weight adjustment", N:0, Timeout:60, launch:BasicWeightAdjustment, MaxRetry:2},
		TestInfo{Desc: "complex.topo", N:0, Timeout:60, launch:ComplexTopo, MaxRetry:2},
		TestInfo{Desc: "star.topo", N:0, Timeout:60, launch:StarTopo, MaxRetry:2},
		TestInfo{Desc: "simple.topo", N:0, Timeout:60, launch:SimpleTopo, MaxRetry:2},
		TestInfo{Desc: "tri.topo", N:0, Timeout:60, launch:TriTopo, MaxRetry:2},
		TestInfo{Desc: "complex2.topo", N:0, Timeout:60, launch:Complex2Topo, MaxRetry:2}}
	
	runTest := func(i int,ti TestInfo) {
		log.Print("Starting Test " + ti.Desc)
		fmt.Printf("Test %d (%s): ",i,ti.Desc)
		err, suc := ti.Execute()
		if err {
			fmt.Println("(ERR)")
		} else {
			fmt.Printf("%t\n", suc)
		}
	}

	if *singleTest != -1 {
		runTest(*singleTest, tests[*singleTest])	
		return
	}
	for i,ti := range tests {
		runTest(i,ti)
	}

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
