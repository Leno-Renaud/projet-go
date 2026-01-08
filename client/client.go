package main

import (
	"io"
	"net"
	"os"
)

func main() {
	conn, _ := net.Dial("tcp", "localhost:9000")
	defer conn.Close()

	in, _ := os.Open("images_sources/asiats_500x500.jpg")
	defer in.Close()

	io.Copy(conn, in)                // envoie l'image
	conn.(*net.TCPConn).CloseWrite() // signale la fin de l'envoi

	out, _ := os.Create("out.png")
	defer out.Close()

	io.Copy(out, conn) // reçoit l'image renvoyée
}
