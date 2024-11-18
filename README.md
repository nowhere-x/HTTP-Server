# HTTP Server Repo for DS Lab1

## HTTP Server
To run/start the server, enter into terminal: go run http-server.go 
To specify which port to use, enter "-port <port number>" into the command line when starting the server
To use the POST method, enter "curl.exe -X POST -F "file=@<filepath/name>" http://localhost:<port>/<outputfilename> into the command line
To use the GET method, enter "curl.exe -X GET http://localhost:<port>/<path_to_file> -o <outfilename/path_to_outputfile>

The HTTP server supports the following file types: ".html", ".txt", ".gif", ".jpeg", ".jpg", ".css".

The scripts intitializes the listener, based on a given port in the command line.
Then it accepts incoming connections, up to 10 concurrent connnections, and a go routine is created to handle each connection.
For each connection the requests are parsed and it is determined wether it is a GET, POST or invalid request. 
The script defaults to /index.html for the root path.
The POST requests results in the files being saved, so it can later be called with a GET request.
The GET requests result in a specified file being displayed or saved.
The code deals with the following ERROR responses:
400	Bad Request (method not supported)
404	File not found
500	Internal server error

## HTTP Proxy

The http-proxy handle a HTTP request from the proxy client in the following pattern:

```
Proxy Client ---(HTTP Request)--> Proxy -----(new HTTP Request)---> HTTP Server
                                                                        |
                                                                        |
Proxy Client <---(new HTTP Response)--- Proxy  <----(HTTP Response)------
```

To achieve concurrent requests handling, the request handling function is called using a GO routine:

```go
maxConnections := 10
nrconns := make(chan struct{}, maxConnections)
for {
    ...
    nrconns <- struct{}{}
    go func() {
        err := handleClient(clientConn, nrconns)
        if err != nil {
            log.Println("Error: ", err)
            logAndRespond(clientConn, fmt.Sprintf("Error handling client: %v", err))
        }
    }()
}

func handleClient(clientConn net.Conn, nrconns chan struct{}) error {
	defer clientConn.Close()
	defer func() {
		<-nrconns
	}()

	...
}
```

```handleClient()``` is the entrance of proxy handling. It will check if the request header is appropriate and accept only when ```GET``` method is used.
```go
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
```

In the ```forwardRequest()```, the proxy will establish a connection with the HTTP server and forward the exact request received from proxy client. If a response is received, the proxy server will send the response back to the proxy client.
```go
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
```

A logging function is used to output error logs and send error messages back to the proxy client:

```go
func logAndRespond(clientConn net.Conn, logMessage string) {
	log.Println(logMessage)
	clientConn.Write([]byte("HTTP/1.1 501 Bad Request\r\nContent-Type: text/plain\r\n\r\n" + logMessage + "\n"))
	clientConn.Close()
}
```


## Compile And Run On A Cloud Server

- Install ```git``` and ```Go``` package on the server:
```sh
# Debian as example
apt install golang git -y
```

- Clone this repository to the cloud server or unzip the downloaded zip file from GitHub/Canvas:
```sh
git clone https://github.com/nowhere-x/HTTP-Server.git
```

- Direct into the http-server folder or http-proxy, use ```go build``` command to build executable binary:
```sh
# build http server
cd ./http-server
go build

# build http proxy
cd ../http-proxy
go build
```

- Run the executable file with specified port. Make sure the server's firewall allows incoming traffic to the port server/proxy is listening on. 
```sh
# run http server
./http-server --port=<your port> &

# run http proxy
./http-proxy --port=<your port> &
```

- Test the server/proxy using a web browser or ```curl```  command:
```sh
# use curl as example

# test server GET
curl -X GET http://example.com:port/
# test server POST
curl -X POST -F "file=@/your_file" http://example.com:port/filename
# test proxy 
curl -x  http://proxyIP:port/ -X GET <target site>
```