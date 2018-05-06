package test

import "os/exec" 
import "os"
import "log"

func (ti *TestInstance) startRouter(router *Router) {
	quitCh := make(chan bool)
	ti.quitChs = append([]chan bool{quitCh}, ti.quitChs...)
	quitChSubRoutine := make(chan bool)

	routerName := router.Name
	cmd := exec.Command("./dr","-port",ti.port,"-name",router.Name)

    outfile, err := os.Create("./" + router.Name + ".txt")
    if err != nil {
        panic(err)
    }

    cmd.Stdout = outfile
    cmd.Stderr = outfile
	cmd.Start()
	go func() {
		defer deferKillUponPanic(cmd)
		<-quitCh
		close(quitChSubRoutine)
		cmd.Process.Kill()
		log.Print("Killed " + routerName)
	}()

	go func() {
		defer deferKillUponPanic(cmd)
		log.Print("Starting " + router.Name)
		err := cmd.Wait()
    	defer outfile.Close()	
		select {
			case <-quitChSubRoutine:
				log.Print(routerName + " terminated")
				return
			default:
		}
		if err != nil {
			panic("router " + routerName + " terminated early " + err.Error())
		}
	}()

}

func (ti *TestInstance) startRouters() {
	ti.lvnsFlushMessages()
	for _,router := range ti.test.Routers {
		ti.startRouter(router)
		msg := ti.lvnsAwaitMsg("connected", []string{"prompt", "port"})
		if msg.Msg != router.Name {
			panic("wrong router name in connected msg got: " + msg.Msg + " expected " + router.Name)
		}
	}


}

