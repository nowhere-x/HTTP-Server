# HTTP Server Repo for DS Lab1
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