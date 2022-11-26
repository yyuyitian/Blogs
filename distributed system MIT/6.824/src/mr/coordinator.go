package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
)

type Coordinator struct {
	// Your definitions here.

}

var filesall []string
var workerIndex int
var mapServers []string
var reduceIndex int

// Your code here -- RPC handlers for the worker to call.

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) SendFiles(args *ExampleArgs, reply *FilesName) error {
	reply.Files = filesall[workerIndex : workerIndex+2]
	reply.Index = workerIndex
	workerIndex += 2
	return nil
}

func (c *Coordinator) SendTasks(args *ExampleArgs, reply *Task) error {
	fmt.Printf("SendTasks!\n")
	reply.Sockets = mapServers
	reply.Index = reduceIndex
	reduceIndex++
	return nil
}

func (c *Coordinator) ReceiveNotify(socket string, reply *Task) error {
	fmt.Printf("ReceiveNotify!\n")
	mapServers = append(mapServers, socket)
	reply.Sockets = mapServers
	return nil
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	fmt.Printf("listen!\n")
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	ret := false

	// Your code here.

	return ret
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// Your code here.
	filesall = files

	c.server()
	return &c
}
