package main

import (
	"fmt"
	"io"
	"net"
)

func main() {
	ln, _ := net.Listen("tcp", ":9000")
	fmt.Println("Serveur démarré sur :9000")
	for {
		conn, _ := ln.Accept()
		go func(c net.Conn) {
			defer c.Close()
			buf, _ := io.ReadAll(c)
			c.Write(buf)
			fmt.Println("Requête complétée")
		}(conn)
	}
}
