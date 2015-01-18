package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
)

type PingRequest struct {
	Host string
}

type PingResponse struct {
	Location string
	Status   string
}

type NetworkNode struct {
	Host string
	Port string
}

func SendNodeRequest(nn NetworkNode, p *PingRequest, response chan PingResponse, wg *sync.WaitGroup) {

	host := nn.Host + ":" + nn.Port
	fmt.Println("Connecting to node: " + host)

	defer wg.Done()

	// Connect to a node
	conn, err := net.Dial("tcp", host)
	defer conn.Close()

	if err != nil {
		log.Print(err.Error())
	}

	// Send the ping request
	encoder := json.NewEncoder(conn)
	encoder.Encode(&p)

	fmt.Println("Request sent to node: " + host)

	// Decode the json response from the node
	var pr PingResponse
	decoder := json.NewDecoder(conn)

	// Blocking read
	if theDisco := decoder.Decode(&pr); theDisco != nil {
		panic(theDisco)
	}

	// Write the json response to the channel from the network node
	response <- pr

	fmt.Println("Response received from node: " + host)
}

func (p *PingRequest) Dispatch() chan PingResponse {

	// Buffer a channel to handle each response
	nodeResponses := make(chan PingResponse, len(networkNodes))

	var wg sync.WaitGroup

	wg.Add(len(networkNodes))

	for _, node := range networkNodes {
		go SendNodeRequest(node, p, nodeResponses, &wg)
	}

	wg.Wait()
	close(nodeResponses)

	fmt.Println("Returning Responses")
	return nodeResponses
}

func handlePingRequest(w http.ResponseWriter, r *http.Request) {
	/**var pingReq PingRequest
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(pingReq); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}**/

	fmt.Println("Receiving ping request from http server.")

	// Temporary
	pingReq := PingRequest{Host: "http://localhost:8080"}

	responses := pingReq.Dispatch()

	var pings []PingResponse
	for node := range responses {
		pings = append(pings, node)
	}

	js, err := json.Marshal(pings)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
