package test

import "net"
import "encoding/binary"
import "gonum.org/v1/gonum/graph"
import "gonum.org/v1/gonum/graph/simple"
import "strconv"


func int2ip(nn uint32) string {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip.String()
}

func Ip2int(ipStr string) uint32 {
	return ip2int(ipStr)
}
func ip2int(ipStr string) uint32 {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		panic("coulndt parse ip" + ipStr)
	}

	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

func NewTest() *Test {
	var initialIP uint32
	initialIP = (0x1) + (0x01<<8) + (0x01<<16) + (0x01<<24)
	return &Test{routerCnt: 0, addrMaskBits: 24, addrIp: initialIP, g: simple.NewWeightedDirectedGraph(0,100), NodeToStruct: make(map[int64]interface{}), Subnets: make(map[SubnetKey]*Subnet), EdgeToInterface: make(map[graph.Edge]*Interface)}
}

type Test struct {
	NodeToStruct map[int64]interface{}
	EdgeToInterface map[graph.Edge]*Interface
	Routers []*Router
	Subnets map[SubnetKey]*Subnet
	routerCnt int
	g *simple.WeightedDirectedGraph
	addrIp uint32
	addrMaskBits uint32
	instance *TestInstance
}
func (t *Test) RouterFromNode(id int64) *Router {
	rt := t.routerFromNode(id)
	if rt == nil {
		panic("expected router")
	}
	return rt
}

func (t *Test) InterfaceFromNode(id int64) *Interface {
	rt := t.interfaceFromNode(id)
	if rt == nil {
		panic("expected interface")
	}
	return rt
}

func (t *Test) SubnetFromNode(id int64) *Subnet {
	rt := t.subnetFromNode(id)
	if rt == nil {
		panic("expected subnet")
	}
	return rt
}

func (t *Test) routerFromNode(id int64) *Router {
	router,ok := t.NodeToStruct[id].(*Router)
	if ok {
		return router
	}
	return nil
}

func (t *Test) subnetFromNode(id int64) *Subnet {
	router,ok := t.NodeToStruct[id].(*Subnet)
	if ok {
		return router
	}
	return nil
}

func (t *Test) interfaceFromNode(id int64) *Interface {
	router,ok := t.NodeToStruct[id].(*Interface)
	if ok {
		return router
	}
	return nil
}

func (t *Test) StartTest() {
	if t.instance != nil {
		panic("test already running")
	}
	ti := NewTestInstance(t)
	ti.Start()
	t.instance = ti
}

func (t *Test) nextRouterName() string {
	t.routerCnt++
	return "d" + strconv.Itoa(t.routerCnt)
}

func (t *Test) nextAddressSpinterface() (uint32, uint32) {
	t.addrIp += (1<<(32-t.addrMaskBits))
	mask := (0xFFFFFFFF)<<(32-t.addrMaskBits)
	return t.addrIp, uint32(mask)
}

func (t *Test) AddCustomRouter(routerName string) *Router {
	n := t.NewNode(routerName)
	t.g.AddNode(n)

	rt := &Router{Name:routerName , Node:n}
	t.Routers = append(t.Routers, rt)

	t.NodeToStruct[n.ID()] = rt

	return rt
}

func (t *Test) AddRouter() *Router {
	routerName := t.nextRouterName()
	return t.AddCustomRouter(routerName)
}

func (t *Test) GetOrCreateSubnet(ip, mask uint32) *Subnet {
	sk := SubnetKey{ip&mask,mask}
	subnet := t.Subnets[sk]
	if subnet != nil {
		return subnet
	}

	n := t.NewNode(int2ip(ip&mask))
	t.g.AddNode(n)
	sn := &Subnet{Address: ip, Mask:mask, Node:n}
	t.Subnets[sk] = sn
	t.NodeToStruct[n.ID()] = sn

	return sn
}

func (t *Test) SetInterfaceCost(i *Interface, cost float64) {
	interfaceToInterfaces := t.NewWeightedEdge(i.Router.Node, i.Node, cost)
	t.g.SetWeightedEdge(interfaceToInterfaces)

	interfaceSubnet := t.NewWeightedEdge(i.Router.Node, i.NodeSubnets, cost)
	t.g.SetWeightedEdge(interfaceSubnet)

	i.Cost = cost
}

func (t *Test) GetInterfaceByIP(ip uint32) *Interface {
	for _,router := range t.Routers {
		for _,intrf := range router.Interfaces {
			if intrf.Address == ip {
				return intrf
			}
		}
	}
	panic("didnt find interface with specified ip")
}

func (t *Test) LinkInterfaces(aIntrf *Interface, bIntrf *Interface) {
	interfaceToInterface := t.NewWeightedEdge(aIntrf.Node, bIntrf.Router.Node, 0)		
	t.g.SetWeightedEdge(interfaceToInterface)	
	t.EdgeToInterface[t.Edge(aIntrf.Node.ID(), bIntrf.Router.Node.ID())] = bIntrf

	interfaceToInterface = t.NewWeightedEdge(bIntrf.Node, aIntrf.Router.Node, 0)		
	t.g.SetWeightedEdge(interfaceToInterface)	
	t.EdgeToInterface[t.Edge(bIntrf.Node.ID(), aIntrf.Router.Node.ID())] = aIntrf
}

func (t *Test) ConnectRoutersUni(a,b *Router, cost float64, ip, mask,toip uint32) {
	subnet := t.GetOrCreateSubnet(ip, mask)

	aIntrf := t.GetOrCreateInterface(a, ip, mask, subnet)	
	bIntrf := t.GetOrCreateInterface(b, toip, mask, subnet)	

	t.SetInterfaceCost(aIntrf, cost)

	interfaceToInterface := t.NewWeightedEdge(aIntrf.Node, b.Node, 0)		
	t.g.SetWeightedEdge(interfaceToInterface)
	t.EdgeToInterface[t.Edge(aIntrf.Node.ID(), b.Node.ID())] = bIntrf

	//t.EdgeToInterfaces[aSubnet.WeightedEdge] = 
}

func (t *Test) nextAddressSpace() (uint32, uint32) {
	t.addrIp += (1<<(32-t.addrMaskBits))
	mask := (0xFFFFFFFF)<<(32-t.addrMaskBits)
	return t.addrIp, uint32(mask)
}
func (t *Test) ConnectRouters(a *Router, b *Router, costAB float64, costBA float64) {
	ip,mask := t.nextAddressSpace()	
	t.ConnectRoutersUni(a,b,costAB, ip, mask, ip+1)
	t.ConnectRoutersUni(b,a,costBA, ip+1, mask, ip)
}

type SubnetKey struct {
	Address uint32
	Mask uint32
}

type Subnet struct {
	Address uint32
	Mask uint32
	Node graph.Node
}

func (s *Subnet) addrSpace() uint32 {
	return s.Address&s.Mask
}

type Router struct {
	Name string
	Interfaces []*Interface	
	Node graph.Node
}

func (t *Test) TearDown() {
	t.instance.TearDown()
}

func (t *Test) MemUsage() []float64 {
	return t.instance.MemUsage()
}

func (r *Router) AddInterface(address, mask uint32, subnet *Subnet, n graph.Node, nodeSubnets graph.Node) (*Interface) {	
	intrf := &Interface{Address: address, Mask:mask, Subnet: subnet, State:"up", Node: n, Router: r, NodeSubnets: nodeSubnets}
	r.Interfaces = append(r.Interfaces, intrf)
	return intrf
}

func (r *Router) GetInterfaceForSubnet(ip, mask uint32) (*Interface) {
	for _,intrf:= range r.Interfaces {
		if (ip&mask) == (intrf.Address&intrf.Mask) && mask == intrf.Mask {
			return intrf
		}
	}
	panic("didnt find interface for subnet (on router)")
}

func (t *Test) GetOrCreateInterface(r* Router, ip, mask uint32, subnet *Subnet) *Interface {
	for _,intrf:= range r.Interfaces {
		if ip == intrf.Address && mask == intrf.Mask {
			return intrf
		}
	}

	n := t.NewNode(r.Name + "_i")
	t.g.AddNode(n)

	nS := t.NewNode(r.Name + "_is")
	t.g.AddNode(nS)

	// interfaceRouter := t.NewWeightedEdge(n, r.Node, 0)	
	// t.g.SetWeightedEdge(interfaceRouter)

	interfaceSubnet := t.NewWeightedEdge(nS, subnet.Node, 0)	
	t.g.SetWeightedEdge(interfaceSubnet)

	intrf := r.AddInterface(ip, mask, subnet, n, nS)
	intrf.Router = r
	t.NodeToStruct[n.ID()] = intrf

	return intrf
}

type InterfacePair struct {
	a *Interface
	b *Interface
}
type Interface struct {
	Address uint32
	Mask uint32
	Cost float64
	State string
	Node graph.Node
	NodeSubnets graph.Node
	Router *Router
	Subnet *Subnet
}

func (i *Interface) MaskBits() int {
	mask := i.Mask
	bits := 0
	for ;mask!=0;mask=mask<<1 {
		bits++
	}
	return bits
}

func (i *Interface) String() string {
	return int2ip(i.Address) + "/" + strconv.Itoa(i.MaskBits()) + ":" + i.StringCost()
}

func (i *Interface) StringCost() string {
	return strconv.FormatFloat(i.Cost, 'f', 0, 64)
}

func (i *Interface) StringIP() string {
	return int2ip(i.Address)
}

