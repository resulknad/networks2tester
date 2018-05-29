package test

import "regexp"
import "log"
import "os/exec" 
import "io"
import "bufio"
import "time"
//import "fmt"


type lvnsMsg struct {
	Type string
	Msg string
	Msg2 string
}

func createChanReader(rd io.Reader, quitCh chan bool) chan string {	
	ch := make(chan string)
	
	go func() {
		scanner := bufio.NewScanner(rd)	
		split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
				advance, token, err = bufio.ScanLines(data, atEOF)
				// prompt marks end of output
				// is w/o newline, therefore we need to extend
				// scanlines
				if advance == 0 && len(data) > 5 && string(data)[:6] == "lvns> " {
					advance = 6
					token = data
					err = nil
				}
				return
		} 
		scanner.Split(split)
		for scanner.Scan() {
		//	fmt.Println(scanner.Text())
			ch<-scanner.Text()
			select {
			case <-quitCh:
				log.Print("chan reader quit")
				return
			default:
			}
		}
		if err := scanner.Err(); err != nil {
			select {
			case <-quitCh:
				log.Print("chan reader quit")
				return
			default:
			}
			log.Print("lvns stdout reading error")
		}
	}()
	return ch
}

func deferKillUponPanic (c *exec.Cmd) {
		if r:=recover();r!=nil {
			log.Print(r)
		}
		c.Process.Kill()
}

func (ti *TestInstance) startLvns() {
	quitChReader := make(chan bool)
	quitChHandler := make(chan bool)

	quitCh := make(chan bool)
	ti.quitChs = append(ti.quitChs,quitCh)
	ti.chLvnsStdout = make(chan lvnsMsg, 10)
	ti.chLvnsStdin = make(chan lvnsMsg, 10)
	cmd := exec.Command("./lvns")

	writer, _ := cmd.StdinPipe()
	reader,_ := cmd.StdoutPipe()
	go ti.lvnsHandler(createChanReader(reader, quitChReader), writer, quitChHandler)
	

	go func() {
		defer deferKillUponPanic(cmd)
		quitSubRoutine := make(chan bool)
		cmd.Start()
		go func() {
			defer deferKillUponPanic(cmd)
			err := cmd.Wait()
			select {
				case <-quitSubRoutine:
					log.Print("lvns terminated")
					return
				default:
			}
			if err != nil {
				panic("lvns crashed: " + err.Error())
			}
		}()
		<-quitCh
		close(quitSubRoutine)
		cmd.Process.Kill()
		log.Print("Killed lvns")
	}()

	select {
	case msg:=<-ti.chLvnsStdout:
		if msg.Type != "port" {
			panic("didnt get port from lvns server upon start")
		}
		ti.port = msg.Msg
	case <-time.After(1*time.Second):
		panic("timeout waiting for port from lvns server")
	}
	log.Print("Started lvns on port " + ti.port)
}

func (ti *TestInstance) lvnsAwaitMsg(mType string, skipMsgs []string) lvnsMsg {
	canSkip := func (mType string) bool {
		for _,m := range skipMsgs {
			if m == mType {
				return true
			}
		}
		return false
	}

	ti.Lock()
	defer ti.Unlock()
	for {
		select {
		case msg:=<-ti.chLvnsStdout:
			if msg.Type != mType && !canSkip(msg.Type) {
				panic("didnt get the expected " + mType + " response from lvns, got " + msg.Type)
			} else if msg.Type == mType {
				return msg
			}
		case <-time.After(5*time.Second):
			panic("timeout waiting for " + mType + " from lvns server")
		}
	}
}
func (ti *TestInstance) SetCost(intrf *Interface) {
	ti.lvnsFlushMessages()
	ti.chLvnsStdin<-lvnsMsg{"write", "cost set intf " + intrf.StringIP() + " " + intrf.StringCost() + "\n",""}
	msg := ti.lvnsAwaitMsg("state", []string{"prompt"})
	if intrf.StringCost() != msg.Msg2 {
		panic("lvns didnt change cost " + msg.Msg2)
	}
	log.Print("lvns server changed cost to " + msg.Msg2)

	ti.lvnsWaitForPrompt()
}

func (ti *TestInstance) SetState(intrf *Interface, state string) {
	ti.chLvnsStdin<-lvnsMsg{"write", "intf " + state + " " + intrf.StringIP() + "\n",""}

	msg := ti.lvnsAwaitMsg("state", []string{"prompt"})
	log.Print("lvns server changed state to " + msg.Msg)
	if msg.Msg != state {
		panic("lvns server didnt respond to state change as expected")
	}
}

func (ti *TestInstance) GetRoutes(routerName string) map[uint32]uint32 {
	ti.Lock()
	defer ti.Unlock()
	ti.lvnsFlushMessages()
	routes := make(map[uint32]uint32)
	ti.chLvnsStdin<-lvnsMsg{"write", "route get " + routerName + " all\n",""}
	
	msg:=<-ti.chLvnsStdout
	if msg.Type != "prompt" {
		panic("expected prompt" + msg.Type)
	}
	for {
		select {
		case msg:=<-ti.chLvnsStdout:
			if msg.Type == "route" {
				routes[ip2int(msg.Msg)] = ip2int(msg.Msg2)
			} else if msg.Type == "prompt" {
				return routes
			}
		case <-time.After(5*time.Second):
			log.Print("didnt get answer upon get route request from lvns")
		}
	}
}

func (ti *TestInstance) lvnsWaitForPrompt() {
	ti.lvnsAwaitMsg("prompt", []string{})
}

func (ti *TestInstance) lvnsFlushMessages() {
	for {
		select {
		case <-ti.chLvnsStdout:
		default:
			return
		}
	}
}

func (ti *TestInstance) lvnsHandler(stdout chan string,stdin io.Writer, quitCh chan bool) {
	defer func() {
		if r:=recover();r!=nil {
			log.Print(r)
		}
	}()
	rPort := regexp.MustCompile("port ([0-9]+)")
	rRoute := regexp.MustCompile(`will route to[\s]+([0-9.]+)[\s]+via a next hop of[\s]+([0-9.]+)[\s]+from eth([0-9])`) 
	rConnected := regexp.MustCompile(`\*\*\*[\s]+([a-zA-Z0-9.]+) is now connected`)
	rInterfaceState := regexp.MustCompile(`Interface state=([a-z]+) cost=([0-9]+)`)

	for {
		select {
		case msg:=<-stdout:
			if msg == "lvns> " {
				ti.chLvnsStdout<-lvnsMsg{"prompt","",""}
			} else if rPort.MatchString(msg) {
				port := rPort.FindAllStringSubmatch(msg, -1)[0][1]
				ti.chLvnsStdout<-lvnsMsg{"port", port,""}
			} else if rRoute.MatchString(msg) {
				matches := rRoute.FindStringSubmatch(msg)
				to := matches[1]
				next_hop := matches[2]
				//intf [3]
				ti.chLvnsStdout<-lvnsMsg{"route", to, next_hop}
			} else if rConnected.MatchString(msg) {
				matches := rConnected.FindStringSubmatch(msg)
				routerName := matches[1]
				ti.chLvnsStdout<-lvnsMsg{"connected", routerName,""}
			} else if rInterfaceState.MatchString(msg) {
				matches := rInterfaceState.FindStringSubmatch(msg)
				state := matches[1]
				cost := matches[2]
				ti.chLvnsStdout<-lvnsMsg{"state", state,cost}
			}
		case msg:=<-ti.chLvnsStdin:
			if msg.Type == "write" {
				//log.Print(msg)
				//log.Print(msg.Msg)
				io.WriteString(stdin, msg.Msg)
			}

		case <-quitCh:
			log.Print("lvns handler quit")
			return
		}
	}
}

