package test

import "io/ioutil"

import "strconv"

import "log"
import "os/exec"

import "gonum.org/v1/gonum/graph"

import "gonum.org/v1/gonum/graph/encoding/dot"
import "gonum.org/v1/gonum/graph/encoding"
import "sort"
type Attrs map[string]string

func (a Attrs) Attributes() []encoding.Attribute {
	var keys []string
	for key := range a {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var attrs []encoding.Attribute
	for _, key := range keys {
		attr := encoding.Attribute{
			Key:   key,
			Value: `"` + a[key] + `"`,
		}
		attrs = append(attrs, attr)
	}
	return attrs
}

type CustomEdge struct {
	graph.WeightedEdge
	Attrs
}

type CustomNode struct {
	graph.Node
	Attrs
}

func (t *Test) Edge(a,b int64) graph.WeightedEdge {
	ce,ok := t.g.Edge(a,b).(CustomEdge)
	if !ok {
		panic("Expected custom edge, didnt get custom edge")
	}
	return ce.WeightedEdge
}


func (t *Test) NewNode(label string) CustomNode {
	cn := CustomNode{}
	cn.Node = t.g.NewNode()
	cn.Attrs = make(Attrs)
	cn.Attrs["label"] = label
	return cn
}


func (t *Test) NewWeightedEdge(a,b graph.Node, weight float64) CustomEdge {
	ce := CustomEdge{}
	ce.WeightedEdge = t.g.NewWeightedEdge(a,b,weight)
	ce.Attrs = make(Attrs)
	ce.Attrs["label"] = strconv.FormatFloat(weight, 'f', 0, 64)
	return ce
}

func (t *Test) DrawGraph(outfile string) {

	dotstring, _ := dot.Marshal(t.g, "", "", "  ")
	ioutil.WriteFile("dotfile", []byte(dotstring), 0644)
	cmd := exec.Command("dot", "-Tsvg", "-o"+outfile)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Print(err)
		stdin.Close()
		return
	}

	stdin.Write([]byte(dotstring))
	stdin.Close()

	cmd.Wait()
}
