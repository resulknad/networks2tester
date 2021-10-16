package test

import "strings"
import "log"
import "sync"

func NewTestInstance(test *Test) (*TestInstance){
	return &TestInstance{test: test}	
}


type TestInstance struct {
	sync.Mutex
	test *Test
	port string
	pids []int
	chLvnsStdout chan lvnsMsg
	chLvnsStdin chan lvnsMsg
	quitChs []chan bool
}

func (ti *TestInstance) TearDown() {
	for _,ch := range ti.quitChs {
		close(ch)	
	}
}

func (ti *TestInstance) createTopology() {
	for _,router := range ti.test.Routers {
		var interfaces []string
		for _,intrf := range router.Interfaces {
			interfaces = append(interfaces,intrf.String())
		}
		ti.chLvnsStdin<-lvnsMsg{"write", "node add " + router.Name + " dr " + strings.Join(interfaces," ") + "\n","", ""}
		ti.lvnsWaitForPrompt()
	}


	added := make(map[InterfacePair]bool)
	
	for _,router := range ti.test.Routers {
		for _,intrf := range router.Interfaces {
			for _,remoteIntrf := range ti.test.getConnectedInterfacesFromInterface(intrf) {

				if added[InterfacePair{intrf,remoteIntrf}] {
					// dont add it twice
					continue
				}

				ti.chLvnsStdin<-lvnsMsg{"write", "link add " + intrf.StringIP() + " " + remoteIntrf.StringIP() + "\n","", ""}
				ti.lvnsWaitForPrompt()
				added[InterfacePair{intrf,remoteIntrf}] = true
				added[InterfacePair{remoteIntrf, intrf}] = true
		}
		//time.Sleep(100000) // 100ms
		}
	}	
	log.Print("Created Topology")
}



func (ti *TestInstance) Start() {
	ti.startLvns()
	ti.createTopology()
	ti.startRouters()
}
