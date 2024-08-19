package node


type Node struct {
	Name string
	Ip string
	Api string
	Memory int
	MemoryAllocated int
	Disk int
	DiskAllocated int
	Role string
	TaskCount int
}

func NewNode(name string, api string, role string) *Node {
	return &Node{
		Name: name,
		Api:  api,
		Role: role,
	}
}