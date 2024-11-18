package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
)

func logAndRespond(clientConn net.Conn, logMessage string) {
	log.Println(logMessage)
	clientConn.Write([]byte("HTTP/1.1 501 Bad Request\r\nContent-Type: text/plain\r\n\r\n" + logMessage + "\n"))
	clientConn.Close()
}

func main() {
	port := flag.String("port", "9001", "port on which the proxy listen")
	flag.Parse()

	proxy, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("Server listen failed: %v\n", err)
		return
	}
	log.Println("Server is listening on port: " + *port)

	maxConnections := 10
	nrconns := make(chan struct{}, maxConnections)
	for {
		clientConn, err := proxy.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err)
			continue
		}
		log.Println("Accepted a new connection from: ", clientConn.RemoteAddr().String())

		nrconns <- struct{}{}
		go func() {
			err := handleClient(clientConn, nrconns)
			if err != nil {
				log.Println("Error: ", err)
				logAndRespond(clientConn, fmt.Sprintf("Error handling client: %v", err))
			}
		}()
	}
}

func handleClient(clientConn net.Conn, nrconns chan struct{}) error {
	defer clientConn.Close()
	defer func() {
		<-nrconns
	}()

	reader := bufio.NewReader(clientConn)
	request, err := http.ReadRequest(reader)
	if err != nil {
		logAndRespond(clientConn, fmt.Sprintf("Error parsing request: %v", err))
		return err
	}

	if request.Method != "GET" {
		logAndRespond(clientConn, fmt.Sprintf("METHOD %s IS NOT SUPPORTED", request.Method))
		return fmt.Errorf("method %s is not supported", request.Method)
	}

	return forwardRequest(clientConn, request)
}

func forwardRequest(clientConn net.Conn, request *http.Request) error {
	client := &http.Client{}
	request.RequestURI = ""
	response, err := client.Do(request)
	if err != nil {
		logAndRespond(clientConn, fmt.Sprintf("Failed to forward request: %v", err))
		return err
	}
	defer response.Body.Close()

	// Write the response back to the client
	if err := response.Write(clientConn); err != nil {
		logAndRespond(clientConn, fmt.Sprintf("Failed to send response to client: %v", err))
		return err
	}

	return nil
}
