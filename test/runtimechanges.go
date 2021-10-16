package test

import "log"

func (t *Test) SetCost(a, b *Router, costAB float64, costBA float64) {
	t.SetCostUni(a,b,costAB)
	t.SetCostUni(b,a,costBA)
}

/*func (t *Test) getRouters(s *Subnet) ([]*Router) {
	panic("deprecated")
	var routers []*Router
	for _, router := range t.Routers {	
		for _,intrf := range router.Interfaces {
			if intrf.Subnet == s {
				//routers = append(routers, intrf.Endpoint)
			}
		}
	}
	return routers
}*/

func (t *Test) getConnectedInterfacesFromInterface(intrf *Interface) []*Interface {
	var interfaces []*Interface

	nodes := t.g.From(intrf.Node.ID())
	for nodes.Next() {
		bRouter := t.routerFromNode(nodes.Node().ID())
		if bRouter != nil {
			interfaces = append(interfaces, t.EdgeToInterface[t.Edge(intrf.Node.ID(), bRouter.Node.ID())])
		}
	}

	return interfaces
}
func (t *Test) getInterfaces(a,b *Router) []*Interface {
	var interfaces []*Interface
	for _,intrf := range a.Interfaces {
		nodes := t.g.From(intrf.Node.ID())
		for nodes.Next() {
			bRouter := t.routerFromNode(nodes.Node().ID())
			if bRouter != nil && bRouter == b {
				interfaces = append(interfaces, intrf)
			}
		}
	}

	if len(interfaces) == 0 {
		panic("Trying to find non existing link...")
	}

	return interfaces
}

func (t *Test) GetSingleInterfaceBetween(a,b *Router) *Interface {
	interfaces := t.getInterfaces(a,b)
	if len(interfaces) > 1 {
		log.Print(interfaces)
		panic("this method doesnt work for multiple connections between routers")
	}
	return interfaces[0]
}
	
func (t *Test) SetCostUni(a,b *Router, costAB float64) {
	intrf := t.GetSingleInterfaceBetween(a,b)
	t.SetCostGraphAndInstance(intrf, costAB)

	log.Printf("Set cost %s -> %s to %d in graph",a.Name, b.Name,costAB)
}

func (t *Test) SetCostGraphAndInstance(a *Interface, costAB float64) {
	t.SetInterfaceCost(a, costAB)
	log.Printf("Set cost to %f in graph",costAB)
	t.instance.SetCost(a)	
}

func (t *Test) TakeDownUni(a,b *Router) {
	intrf := t.GetSingleInterfaceBetween(a,b)
	t.TakeDownInterface(intrf)
}

func (t *Test) TakeDownInterface(intrf *Interface) {	
	t.g.RemoveEdge(intrf.Router.Node.ID(), intrf.Node.ID())
	t.g.RemoveEdge(intrf.Router.Node.ID(), intrf.NodeSubnets.ID())
	log.Printf("Removing edge %s -> (interface) in graph",intrf.Router.Name)

	if intrf.State == "down" {
		panic("Taking down interface which is already down")
	}

	t.instance.SetState(intrf, "down")
	intrf.State = "down"
}

func (t *Test) PutUpInterface(intrf *Interface) {	
	log.Printf("Adding edge %s -> (interface) in graph",intrf.Router.Name)
	t.SetInterfaceCost(intrf,intrf.Cost)

	if intrf.State == "up" {
		panic("Putting up interface which is already up")
	}
	t.instance.SetState(intrf, "up")
	intrf.State = "up"
}

func (t *Test) PutUpUni(a,b *Router) {
	intrf := t.GetSingleInterfaceBetween(a,b)
	t.PutUpInterface(intrf)
}
