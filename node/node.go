package node
 
type Node struct {
    Name            string
    Ip              string
    Cores           int
    Memory          int
    MemoryAllocated int
    Disk            int
    DiskAllocated   int
    Role            string
    TaskCount       int
}

/*
A node is an object that represents any machine in our cluster. 
For example, the manager is one type of node.
The worker, of which there can be more than one, is another type of node. 
The manager will make extensive use of node objects to represent workers.
*/