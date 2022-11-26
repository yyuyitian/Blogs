package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sort"
	"strconv"
)

type WorkerSocket struct {
	// Your definitions here.

}

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

var filesToDeal []string
var fileNameglobal string
var serverName string

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string, role string) {
	if role == "map" {
		files := RequstFiles()
		intermediate := []KeyValue{}
		for _, filename := range files.Files {
			fmt.Println("open file" + filename)
			file, err := os.Open(filename)
			if err != nil {
				log.Fatalf("cannot open %v", filename)
			}
			content, err := ioutil.ReadAll(file)
			if err != nil {
				log.Fatalf("cannot read %v", filename)
			}
			file.Close()
			kva := mapf(filename, string(content))
			intermediate = append(intermediate, kva...)
		}
		sort.Sort(ByKey(intermediate))
		indexStr := strconv.Itoa(files.Index)
		oname := "mr-out-" + indexStr + "-"
		fileNameglobal = oname
		ofile, _ := os.Create(oname)

		//
		// call Reduce on each distinct key in intermediate[],
		// and print the result to mr-out-0.
		//
		i := 0
		ofilevar := ofile
		encvar := json.NewEncoder(ofilevar)

		aflag := false
		hflag := false
		oflag := false
		uflag := false
		for i < len(intermediate) {
			word := intermediate[i].Key
			if word[0] == 'A' && aflag == false {
				fmt.Println("meet a files")
				ofile1, _ := os.Create(oname + "-A")
				encvar1 := json.NewEncoder(ofile1)
				encvar = encvar1
				aflag = true
			} else if word[0] == 'H' && hflag == false {
				ofilevar.Close()
				fmt.Println("meet h files")
				ofile2, _ := os.Create(oname + "-H")
				encvar2 := json.NewEncoder(ofile2)
				encvar = encvar2
				hflag = true
			} else if word[0] == 'O' && oflag == false {
				ofilevar.Close()
				fmt.Println("meet o files")
				ofile3, _ := os.Create(oname + "-O")
				encvar3 := json.NewEncoder(ofile3)
				encvar = encvar3
				oflag = true
			} else if word[0] == 'U' && uflag == false {
				ofilevar.Close()
				fmt.Println("meet u files")
				ofile4, _ := os.Create(oname + "-U")
				encvar4 := json.NewEncoder(ofile4)
				encvar = encvar4
				uflag = true
			}
			// j := i + 1
			// for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
			// 	j++
			// }
			// values := []string{}
			// for k := i; k < j; k++ {
			// 	values = append(values, intermediate[k].Value)
			// }
			// output := reducef(intermediate[i].Key, values)
			//this is the correct format for each line of Reduce output.
			kv := intermediate[i]
			err := encvar.Encode(&kv)
			if err != nil {
				log.Fatalf("write failed")
			}
			i++
		}
		ofile.Close()
		notifyDone()
		// uncomment to send the Example RPC to the coordinator.
		// w := WorkerSocket{}
		// sockname := workerSock()
		// serverName = sockname
		// w.server()
		// for w.Done() == false {
		// 	time.Sleep(time.Second)
		// }
	} else if role == "reduce" {
		task := RequestTask()
		index := task.Index
		log.Println("task index is" + strconv.Itoa(index))
		var letter string
		if index == 0 {
			letter = "A"
		} else if index == 1 {
			letter = "H"
		} else if index == 2 {
			letter = "O"
		} else if index == 3 {
			letter = "U"
		}
		kva := []KeyValue{}
		v := 0
		for v <= 6 {
			filename := "mr-out-" + strconv.Itoa(v) + "--" + letter
			file, err := os.Open(filename)
			if err != nil {
				log.Fatalf("cannot open %v", filename)
			}
			dec := json.NewDecoder(file)
			fmt.Println("NewDecoder")
			for {
				var kv KeyValue
				if err := dec.Decode(&kv); err != nil {
					break
				}
				kva = append(kva, kv)
			}
			fmt.Println("NewDecoder finish")
			if err != nil {
				log.Fatalf("cannot read %v", filename)
			}
			file.Close()
			v += 2
		}
		fmt.Println("key is:" + kva[0].Key + kva[0].Value)
		reducefile, _ := os.Create("reduce" + strconv.Itoa(index))
		i := 0
		for i < len(kva) {
			j := i + 1
			for j < len(kva) && kva[j].Key == kva[i].Key {
				j++
			}
			values := []string{}
			for k := i; k < j; k++ {
				values = append(values, kva[k].Value)
			}
			output := reducef(kva[i].Key, values)
			fmt.Fprintf(reducefile, "%v %v\n", kva[i].Key, output)
			i = j + 1
		}
	}
}

// start a thread that listens for RPCs from worker.go
func (w *WorkerSocket) server() {
	rpc.Register(w)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	// os.Remove(sockname)
	l, e := net.Listen("unix", serverName)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	fmt.Printf("listen!\n")
	go http.Serve(l, nil)
}

func (w *WorkerSocket) Done() bool {
	ret := false

	// Your code here.

	return ret
}

// send files to reduce worker
func (w *WorkerSocket) SendfilesToReduce(index int, reply *MapResult) error {

	return nil
}

func RequestFilesFromReduce(index int, server string) []KeyValue {
	reply := MapResult{}
	fmt.Printf("RequestFilesFromReduce!\n")
	ok := callMap("WorkerSocket.SendfilesToReduce", &index, &reply, server)
	if ok {
		fmt.Printf("RequestFilesFromReduce call success!\n" + strconv.Itoa(len(reply.Kvs)))
		return reply.Kvs
	} else {
		fmt.Printf("call failed!\n")
		return nil
	}
}

// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

func RequstFiles() *FilesName {
	args := ExampleArgs{}
	reply := FilesName{}
	ok := call("Coordinator.SendFiles", &args, &reply)
	if ok {
		fmt.Println(reply.Files)
	} else {
		fmt.Printf("call failed!\n")
	}
	return &reply
}

func RequestTask() *Task {
	args := ExampleArgs{}
	reply := Task{}
	ok := call("Coordinator.SendTasks", &args, &reply)
	if ok {
		fmt.Println(reply.Sockets)
		fmt.Println(reply.Index)
	} else {
		fmt.Printf("call failed!\n")
	}
	return &reply
}

func notifyDone() {
	args := serverName
	reply := Task{}
	ok := call("Coordinator.ReceiveNotify", &args, &reply)
	if ok {
		fmt.Println(reply.Sockets)
	} else {
		fmt.Printf("call failed!\n")
	}
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	fmt.Println("dialing")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}
	fmt.Printf("call failed has error!\n")
	fmt.Println(err)
	return false
}

func callMap(rpcname string, args interface{}, reply interface{}, server string) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := server
	fmt.Printf("callMap" + sockname)
	c, err := rpc.DialHTTP("unix", sockname)
	fmt.Println("dialing")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()
	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}
	fmt.Printf("call failed has error!\n")
	fmt.Println(err)
	return false
}
