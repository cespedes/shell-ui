package main

import (
	"fmt"
	"sync"
	"bytes"
	"log"
	"time"
	"os"
	"os/exec"
	"net"
	"flag"
//	"syscall"
)

func execUsage() {
	fmt.Fprintln(os.Stderr, `Usage: mui exec [-shell interpreter] script [args]

Use "mui help exec" for more information`)
}

func net_listen() {
	ln, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	addr := ln.Addr()
	tcpaddr, err := net.ResolveTCPAddr(addr.Network(), addr.String())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Listening on port %d\n", tcpaddr.Port)
	go func() {
		buf := make([]byte, 128)
		for {
			conn, _ := ln.Accept()
			n, _ := buf_stdout.Read(buf)
			if n > 0 {
				log.Printf("sending %d bytes\n", n)
				conn.Write(buf[:n])
			}
			conn.Close()
		}
	}()
}

type Buffer struct {
    b bytes.Buffer
    m sync.Mutex
}
func (b *Buffer) Read(p []byte) (n int, err error) {
    b.m.Lock()
    defer b.m.Unlock()
    return b.b.Read(p)
}
func (b *Buffer) Write(p []byte) (n int, err error) {
    b.m.Lock()
    defer b.m.Unlock()
    return b.b.Write(p)
}

var buf_stdout Buffer

func executeScript(shell string, args []string) {
	var err error
	cmd := exec.Command(shell, args...)
//	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	pipe_stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
/*
	cmd.Stdout = &buf_stdout
*/
//	r, w, err := os.Pipe()
//	cmd.Stdout = w
	fmt.Printf("Executing: %+v\n", cmd)
	if err = cmd.Start(); err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, 128)
	for {
		n, err := pipe_stdout.Read(buf)
		log.Printf("read %d bytes from command", n)
		if n > 0 {
			buf_stdout.Write(buf[:n])
			log.Printf("wrote %d bytes into buffer", n)
		} else {
			log.Print(err)
			break
		}
	}
/*
	for {
		out := make([]byte, 1024)
		n, err := stdout.Read(out)
		if err != nil {
			break
		}
		fmt.Printf("%s", out[:n])
	}
*/
/*
	if err = cmd.Wait(); err != nil {
		log.Fatal(err)
	}
*/

/*
	ch := make(chan error)
	go func() {
		ch <- cmd.Run()
	}()
	select {
		case err := <- ch:
			fmt.Printf("Error: %v\n", err)
	}
	close(ch)
*/
}

func runExec(args []string) {
	f := flag.NewFlagSet("mui exec", flag.ContinueOnError)
	f.Usage = execUsage
	pshell := f.String("shell", "/bin/sh", "Interpreter to use")
	f.Parse(args)
	args = f.Args()
	fmt.Printf("shell = %v\n", *pshell)
	if len(args) < 1 {
		execUsage()
		os.Exit(1)
	}
	path, err := exec.LookPath(args[0])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("path = %s\n", path)
	net_listen()
	executeScript(*pshell, args)
	for {
		log.Printf("tick\n")
		time.Sleep(1 * time.Second)
	}
}
