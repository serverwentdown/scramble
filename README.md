
# scramble

A simple tool to perform XOR as a TCP proxy. 

## Usage 

```
$ ./scramble -help
Usage of ./scramble:
  -connect string
    	forward to ip and port (default ":8080")
  -key int
    	key to xor the data (default 170)
  -listen string
    	listen on ip and port (default ":8081")
```

## Use with a SOCKS proxy

This tool may come really useful when trying to bypass filters that perform packet inspection. After starting a SOCKS proxy listening on the server, `scramble` can connect to the proxy and listen on an exposed port to provide "obscured" SOCKS proxying if `scramble` is also run on the client. 
