package bench

import (
	//"fmt"

	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"testing"

	"github.com/hprose/hprose-go"
	hproserpc "github.com/hprose/hprose-golang/rpc"
)

// BenchmarkHprose2 is ...
func BenchmarkHprose2(b *testing.B) {
	b.StopTimer()
	server := hproserpc.NewTCPServer("")
	server.AddFunction("hello", hello, hproserpc.Options{})
	server.Handle()
	client := hproserpc.NewTCPClient(server.URI())
	var ro *RO
	client.UseService(&ro)
	defer server.Close()
	defer client.Close()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ro.Hello("World")
	}
	b.StopTimer()
}

// BenchmarkHprose2Unix is ...
func BenchmarkHprose2Unix(b *testing.B) {
	b.StopTimer()
	server := hproserpc.NewUnixServer("")
	server.AddFunction("hello", hello, hproserpc.Options{})
	server.Handle()
	client := hproserpc.NewUnixClient(server.URI())
	var ro *RO
	client.UseService(&ro)
	defer server.Close()
	defer client.Close()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ro.Hello("World")
	}
	b.StopTimer()
}

// BenchmarkHprose is ...
func BenchmarkHprose(b *testing.B) {
	b.StopTimer()
	server := hprose.NewTcpServer("")
	server.AddFunction("hello", hello)
	server.Handle()
	client := hprose.NewClient(server.URL)
	var ro *RO
	client.UseService(&ro)
	defer server.Stop()
	// result, _ := ro.Hello("World")
	// fmt.Println(result)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ro.Hello("World")
	}
	b.StopTimer()
}

// BenchmarkHproseUnix is ...
func BenchmarkHproseUnix(b *testing.B) {
	b.StopTimer()
	server := hprose.NewUnixServer("")
	server.AddFunction("hello", hello)
	server.Handle()
	client := hprose.NewClient(server.URL)
	var ro *RO
	client.UseService(&ro)
	defer server.Stop()
	// result, _ := ro.Hello("World")
	// fmt.Println(result)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ro.Hello("World")
	}
	b.StopTimer()
}

// BenchmarkGobRPC is ...
func BenchmarkGobRPC(b *testing.B) {
	b.StopTimer()
	server := rpc.NewServer()
	server.Register(new(Hello))
	listener, _ := net.Listen("tcp", "")
	defer listener.Close()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.ServeConn(conn)
		}
	}()
	client, _ := rpc.Dial("tcp", listener.Addr().String())
	defer client.Close()
	var args = &Args{"World"}
	var reply string
	// client.Call("Hello.Hello", &args, &reply)
	// fmt.Println(reply)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		client.Call("Hello.Hello", &args, &reply)
	}
	b.StopTimer()
}

// BenchmarkGobRPCUnix is ...
func BenchmarkGobRPCUnix(b *testing.B) {
	b.StopTimer()
	server := rpc.NewServer()
	server.Register(new(Hello))
	listener, _ := net.Listen("unix", "/tmp/gobrpc.sock")
	defer listener.Close()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.ServeConn(conn)
		}
	}()
	client, _ := rpc.Dial("unix", "/tmp/gobrpc.sock")
	defer client.Close()
	var args = &Args{"World"}
	var reply string
	// client.Call("Hello.Hello", &args, &reply)
	// fmt.Println(reply)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		client.Call("Hello.Hello", &args, &reply)
	}
	b.StopTimer()
}

// BenchmarkJSONRPC is ...
func BenchmarkJSONRPC(b *testing.B) {
	b.StopTimer()
	server := rpc.NewServer()
	server.Register(new(Hello))
	listener, _ := net.Listen("tcp", "")
	defer listener.Close()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	}()
	client, _ := jsonrpc.Dial("tcp", listener.Addr().String())
	defer client.Close()
	var args = &Args{"World"}
	var reply string
	// client.Call("Hello.Hello", &args, &reply)
	// fmt.Println(reply)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		client.Call("Hello.Hello", &args, &reply)
	}
	b.StopTimer()
}

// BenchmarkJSONRPCUnix is ...
func BenchmarkJSONRPCUnix(b *testing.B) {
	b.StopTimer()
	server := rpc.NewServer()
	server.Register(new(Hello))
	listener, _ := net.Listen("unix", "/tmp/jsonrpc.sock")
	defer listener.Close()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	}()
	client, _ := jsonrpc.Dial("unix", "/tmp/jsonrpc.sock")
	defer client.Close()
	var args = &Args{"World"}
	var reply string
	// client.Call("Hello.Hello", &args, &reply)
	// fmt.Println(reply)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		client.Call("Hello.Hello", &args, &reply)
	}
	b.StopTimer()
}
