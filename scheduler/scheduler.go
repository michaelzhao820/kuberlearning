package scheduler

import (
	"kuberlearning/node"
	"kuberlearning/task"
)

type Scheduler interface {
	SelectCanidateNodes(t task.Task, nodes []*node.Node) []*node.Node
	Score(t task.Task, nodes []*node.Node) map[string]float64
	Pick(scores map[string]float64, canidates []*node.Node) *node.Node
}

type RoundRobin struct {
	Name string
	LastWorker int
}

func (r *RoundRobin) SelectCanidateNodes(t task.Task, nodes []*node.Node)[]*node.Node {
	return nodes
}

func (r *RoundRobin) Score(t task.Task, nodes []*node.Node) map[string]float64{

	scoreMap := make(map[string]float64)
	var newWorker int
	if r.LastWorker+1 < len(nodes) {
		newWorker = r.LastWorker + 1
        r.LastWorker+=1
    }else{
		newWorker = 0
        r.LastWorker = 0
    }
	for index,node := range(nodes) {
		if index == newWorker{
			scoreMap[node.Name] = 0.1
		}else{
			scoreMap[node.Name] = 1.0
		}
	}
	return scoreMap
}

func (r *RoundRobin) Pick(scores map[string]float64, canidates []*node.Node) *node.Node {
	var returnNode *node.Node = canidates[0]
	var lowestScore float64 = scores[returnNode.Name]
	for _,node := range (canidates) {
		if scores[node.Name] < lowestScore{
			lowestScore = scores[node.Name]
			returnNode = node
		}
	}
	return returnNode

}