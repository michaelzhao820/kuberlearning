package scheduler

import (
	"kuberlearning/node"
	"kuberlearning/task"
	"log"
	"math"
	"time"
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

type Epvm struct {
	Name string
}
const (
	// LIEB square ice constant
	// https://en.wikipedia.org/wiki/Lieb%27s_square_ice_constant
	LIEB = 1.53960071783900203869
)

func (e *Epvm) SelectCandidateNodes(t task.Task, nodes []*node.Node) []*node.Node {
	return selectCandidateNodes(t, nodes)
}

func (e *Epvm) Score(t task.Task, nodes []*node.Node) map[string]float64 {
	nodeScores := make(map[string]float64)
	maxJobs := 4.0

	for _, node := range nodes {
		cpuUsage, err := calculateCpuUsage(node)
		if err != nil {
			log.Printf("error calculating CPU usage for node %s, skipping: %v\n", node.Name, err)
			continue
		}
		cpuLoad := calculateLoad(*cpuUsage, math.Pow(2, 0.8))

		memoryAllocated := float64(node.Stats.MemUsedKb()) + float64(node.MemoryAllocated)
		memoryPercentAllocated := memoryAllocated / float64(node.Memory)

		newMemPercent := (calculateLoad(memoryAllocated+float64(t.Memory/1000), float64(node.Memory)))
		memCost := math.Pow(LIEB, newMemPercent) + math.Pow(LIEB, (float64(node.TaskCount+1))/maxJobs) - math.Pow(LIEB, memoryPercentAllocated) - math.Pow(LIEB, float64(node.TaskCount)/float64(maxJobs))
		cpuCost := math.Pow(LIEB, cpuLoad) + math.Pow(LIEB, (float64(node.TaskCount+1))/maxJobs) - math.Pow(LIEB, cpuLoad) - math.Pow(LIEB, float64(node.TaskCount)/float64(maxJobs))

		nodeScores[node.Name] = memCost + cpuCost
	}
	return nodeScores
}

func (e *Epvm) Pick(scores map[string]float64, candidates []*node.Node) *node.Node {
	minCost := 0.00
	var bestNode *node.Node
	for idx, node := range candidates {
		if idx == 0 {
			minCost = scores[node.Name]
			bestNode = node
			continue
		}

		if scores[node.Name] < minCost {
			minCost = scores[node.Name]
			bestNode = node
		}
	}
	return bestNode
}

func selectCandidateNodes(t task.Task, nodes []*node.Node) []*node.Node {
	var candidates []*node.Node
	for node := range nodes {

		if checkDisk(t, nodes[node].Disk-nodes[node].DiskAllocated) {
			candidates = append(candidates, nodes[node])
		}

	}

	return candidates
}

func checkDisk(t task.Task, diskAvailable int64) bool {
	return t.Disk <= diskAvailable
}

func calculateLoad(usage float64, capacity float64) float64 {
	return usage / capacity
}

// https://stackoverflow.com/questions/23367857/accurate-calculation-of-cpu-usage-given-in-percentage-in-linux
func calculateCpuUsage(node *node.Node) (*float64, error) {
	stat1, err := node.GetStats()
	if err != nil {
		return nil, err
	}
	time.Sleep(3 * time.Second)
	stat2, err := node.GetStats()
	if err != nil {
		return nil, err
	}

	stat1Idle := stat1.CpuStats.Idle + stat1.CpuStats.IOWait
	stat2Idle := stat2.CpuStats.Idle + stat2.CpuStats.IOWait

	stat1NonIdle := stat1.CpuStats.User + stat1.CpuStats.Nice + stat1.CpuStats.System + stat1.CpuStats.IRQ + stat1.CpuStats.SoftIRQ + stat1.CpuStats.Steal
	stat2NonIdle := stat2.CpuStats.User + stat2.CpuStats.Nice + stat2.CpuStats.System + stat2.CpuStats.IRQ + stat2.CpuStats.SoftIRQ + stat2.CpuStats.Steal

	stat1Total := stat1Idle + stat1NonIdle
	stat2Total := stat2Idle + stat2NonIdle

	total := stat2Total - stat1Total
	idle := stat2Idle - stat1Idle

	var cpuPercentUsage float64
	if total == 0 && idle == 0 {
		cpuPercentUsage = 0.00
	} else {
		cpuPercentUsage = (float64(total) - float64(idle)) / float64(total)
	}
	return &cpuPercentUsage, nil
}




