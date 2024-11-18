package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	port := flag.String("port", "8080", "Port on which the server listens")
	flag.Parse()
	ln, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		fmt.Println(err)
		return
	}

	nrconns := make(chan struct{}, 10)
	fmt.Println("Server is listening on port :" + *port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		fmt.Println("Accepted a new connection")

		nrconns <- struct{}{}
		go GoRoutineConn(conn, nrconns)

	}
}

func GoRoutineConn(conn net.Conn, nrconns chan struct{}) {
	defer conn.Close()
	defer func() { <-nrconns }()
	err := processRequest(conn)
	if err != nil {
		fmt.Println("Error processing request:", err)
	}
}

func processRequest(conn net.Conn) error {
	reader := bufio.NewReader(conn)
	request, err := http.ReadRequest(reader)
	fmt.Println(request, err)
	if err != nil {
		fmt.Println("Error parsing request:", err)
		sendResponse(conn, http.StatusBadRequest, "text/plain", strings.NewReader("400 Bad Request - Invalid Request Format"))
		return fmt.Errorf("invalid request format: %v", err)
	}

	// Check for invalid or unsupported HTTP methods
	switch request.Method {
	case "GET":
		return handleGet(conn, request)
	case "POST":
		return handlePost(conn, request)
	default:
		// Respond with a 400 for unsupported methods or malformed requests
		fmt.Println("Error 400: Bad Request - Unsupported Method")
		sendResponse(conn, http.StatusBadRequest, "text/plain", strings.NewReader("400 Bad Request - Unsupported Method"))
		return fmt.Errorf("unsupported method: %s", request.Method)
	}
}

func handleGet(conn net.Conn, request *http.Request) error {
	path := "." + request.URL.Path
	if path == "./" || path == "." || path == "./favicon.ico" || path == "./meta.json" {
		path = "./index.html" // Default to `index.html` for root path
	}

	if !checkExtension(path) {
		fmt.Println("Error 400: Bad Request - Invalid file extension")
		sendResponse(conn, http.StatusBadRequest, "text/plain", strings.NewReader("400 Bad Request: Invalid file extension"))
		return fmt.Errorf("invalid file extension")
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Error 404: File not found")
			sendResponse(conn, http.StatusNotFound, "text/plain", strings.NewReader("Error 404: File not found"))
		} else {
			fmt.Println("Error 500: Server Error")
			sendResponse(conn, http.StatusInternalServerError, "text/plain", strings.NewReader("500 Internal Server Error"))
		}
		return err
	}
	defer file.Close()

	// Get the content type for the requested file.
	contentType := getType(path)

	// Send the file content as the response.
	sendResponse(conn, http.StatusOK, contentType, file)
	return nil
}

func handlePost(conn net.Conn, request *http.Request) error {

	// Retrieve the file from the form (assuming the form field name is "file")
	file, _, err := request.FormFile("file")
	if err != nil {
		fmt.Println("Error retrieving file from form:", err)
		sendResponse(conn, http.StatusBadRequest, "text/plain", strings.NewReader("400 Bad Request - No file found"))
		return err
	}
	defer file.Close()

	// Create a file at the specified path (use the requested path, like "/dogtest.jpg")
	path := "." + request.URL.Path
	outFile, err := os.Create(path)
	if err != nil {
		fmt.Println("Error creating file:", err)
		sendResponse(conn, http.StatusInternalServerError, "text/plain", strings.NewReader("500 Internal Server Error"))
		return err
	}
	defer outFile.Close()

	// Copy the uploaded file content to the new file
	_, err = io.Copy(outFile, file)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		sendResponse(conn, http.StatusInternalServerError, "text/plain", strings.NewReader("500 Internal Server Error"))
		return err
	}

	// Send a 201 Created response
	sendResponse(conn, http.StatusCreated, "text/plain", strings.NewReader("Created"))
	return nil
}

func checkExtension(path string) bool {
	extlist := []string{".html", ".txt", ".gif", ".jpeg", ".jpg", ".css"}
	ext := strings.ToLower(filepath.Ext(path))
	for _, allowed := range extlist {
		if ext == allowed {
			return true
		}
	}
	return false
}

func getType(path string) string {
	switch {
	case strings.HasSuffix(path, ".html"):
		return "text/html"
	case strings.HasSuffix(path, ".txt"):
		return "text/plain"
	case strings.HasSuffix(path, ".gif"):
		return "image/gif"
	case strings.HasSuffix(path, ".jpeg"), strings.HasSuffix(path, ".jpg"):
		return "image/jpeg"
	case strings.HasSuffix(path, ".css"):
		return "text/css"
	default:
		fmt.Printf("Error: Unrecognized file extension for path: %s\n", path)
		return "application/octet-stream"
	}
}

func sendResponse(conn net.Conn, statusCode int, contentType string, body io.Reader) {
	writer := bufio.NewWriter(conn)

	// Send the status line and headers
	statusText := http.StatusText(statusCode)
	fmt.Fprintf(writer, "HTTP/1.1 %d %s\r\n", statusCode, statusText)
	fmt.Fprintf(writer, "Content-Type: %s\r\n", contentType)
	fmt.Fprintf(writer, "Connection: close\r\n")
	fmt.Fprintf(writer, "\r\n")

	// Send the body content
	if body != nil {
		io.Copy(writer, body)
	}
	writer.Flush() // Ensure the content is actually sent
}
