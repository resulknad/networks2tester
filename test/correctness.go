package test

import "time"
import "gonum.org/v1/gonum/graph/path" 
import "log"
func (t *Test) WaitUntilCorrect(timeoutSec int) bool {
	approxElapsed := 0
	for !t.ShortestRoutesCorrect() {
		time.Sleep(1*time.Second)	
		approxElapsed += 1
		if approxElapsed >= timeoutSec {
			return false
		}
	}
	return true
}

func (t *Test) ShortestRoutesCorrect() bool {
	shortestPaths := path.DijkstraAllPaths(t.g)
	for _,router := range t.Routers {
		drRoutes := t.instance.GetRoutes(router.Name)	
		
		//log.Printf("Got %d, total subnets %d",  len(drRoutes), len(t.Subnets))

		for _,subnet := range t.Subnets {
			correctForThisSubnet := false
			nextHop := drRoutes[subnet.Address&subnet.Mask]
			var nextHopCorrect uint32
			shortestToSubnet,weight := shortestPaths.AllBetween(router.Node.ID(), subnet.Node.ID())
			// if shortest path > 16, there is no path
			if weight>=16 {
				if nextHop == 0xFFFFFFFF {
					correctForThisSubnet = true
				} else {
					log.Printf("Path from %s (%d) to subnet (%s) shouldnt exist", router.Name, router.Node.ID(), int2ip(subnet.addrSpace()))
					log.Print("Shouldnt have a path, but it does")
					return false
				}
			}

			for _, nodesList := range shortestToSubnet {
				switch len(nodesList) {
					case 2:
						panic("Route of length 2 doesnt make sense (Router -> Interface)")
					case 3:
						subn := t.SubnetFromNode(nodesList[2].ID())
						if subn != subnet {
							panic("Route to wrong subnet")
						}
						nextHopCorrect = 0
					default:
						nextInterface := t.EdgeToInterface[t.Edge(nodesList[1].ID(),nodesList[2].ID())]
						nextHopCorrect = nextInterface.Address
				}

				if nextHop ==  nextHopCorrect {
					// log.Printf("For route from %s to subnet %s (%s), got next hop %s, should be %s", router.Name, int2ip(subnet.Address), int2ip(subnet.Mask), int2ip(nextHop), int2ip(nextHopCorrect))
					correctForThisSubnet = true
					break
				}
			}
			if !correctForThisSubnet {
				//routersOnSubnet := t.getRouters(subnet)
				log.Printf("Path from %s (%d) to subnet (%s) isnt ok, weight in graph: %f", router.Name, router.Node.ID(), int2ip(subnet.addrSpace()), weight)
				return false
			}
		}	
	}
	return true
}
