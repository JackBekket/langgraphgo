package state_graph

import (
	"context"
	"errors"
	"fmt"
)

type StateGraph struct {
    nodes map[string]Node
    edges []Edge
    entryPoint string
    state interface{}
}

type Node struct {
    Name string
    Function func(ctx context.Context, state interface{}) (interface{}, error)
}

type Edge struct {
    From string
    To func(ctx context.Context, state interface{}) string
}

const END = "END"

var (
    ErrEntryPointNotSet = errors.New("entry point not set")
    ErrNodeNotFound = errors.New("node not found")
    ErrNoOutgoingEdge = errors.New("no outgoing edge found for node")
)

func NewStateGraph() *StateGraph {
    return &StateGraph{
        nodes: make(map[string]Node),
    }
}

func (g *StateGraph) AddNode(name string, fn func(ctx context.Context, state interface{}) (interface{}, error)) {
    g.nodes[name] = Node{
        Name: name,
        Function: fn,
    }
}

func (g *StateGraph) AddEdge(from, to string) {
    g.edges = append(g.edges, Edge{
        From: from,
        To: func(_ context.Context, state interface{}) string {
            return to
        },
    })
}

func (g *StateGraph) AddConditionalEdge(from string, condition func(ctx context.Context, state interface{}) string) {
    g.edges = append(g.edges, Edge{
        From: from,
        To: condition,
    })
}

func (g *StateGraph) SetEntryPoint(name string) {
    g.entryPoint = name
}

type Runnable struct {
    graph *StateGraph
}

func (g *StateGraph) Compile() (*Runnable, error) {
    if g.entryPoint == "" {
        return nil, ErrEntryPointNotSet
    }

    return &Runnable{
        graph: g,
    }, nil
}

func (r *Runnable) Invoke(ctx context.Context, initialState interface{}) (interface{}, error) {
    currentNode := r.graph.entryPoint
    state := initialState

    for {
        if currentNode == END {
            break
        }

        node, ok := r.graph.nodes[currentNode]
        if !ok {
            return nil, fmt.Errorf("%w: %s", ErrNodeNotFound, currentNode)
        }

        var err error
        state, err = node.Function(ctx, state)
        if err != nil {
            return nil, fmt.Errorf("error in node %s: %w", currentNode, err)
        }

        foundNext := false
        for _, edge := range r.graph.edges {
            if edge.From == currentNode {
                currentNode = edge.To(ctx, state)
                foundNext = true
                break
            }
        }

        if !foundNext {
            return nil, fmt.Errorf("%w: %s", ErrNoOutgoingEdge, currentNode)
        }
    }

    return state, nil
}
