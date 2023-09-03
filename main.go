package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)
type Server interface{
	address() string
	isAlive() bool
	Serve(r http.ResponseWriter,w *http.Request)
}
type simpleServer struct{
	addr string
	proxy *httputil.ReverseProxy
}

func newSimpleServer(addr string) *simpleServer{
	serverUrl,err:= url.Parse(addr)
	handleErr(err)
	return &simpleServer{
		addr:addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

type LoadBalancer struct{
	port string
	RoundRobbinCount int
	servers []Server
}

func newLoadBalancer(port string,servers []Server ) * LoadBalancer{
	return &LoadBalancer{
		port: port,
		RoundRobbinCount: 0,
		servers: servers ,
	}
}

func handleErr(err error){
	if err !=nil{
		fmt.Printf("error%v\n", err)
		os.Exit(1)
	}
}

func (s *simpleServer) address() string {return s.addr}

func (s *simpleServer) isAlive() bool{return true}

func (s *simpleServer) Serve(r http.ResponseWriter, w *http.Request){
	s.proxy.ServeHTTP(r,w)
}

func (lb* LoadBalancer) getNextAvailableServer() Server{
	server := lb.servers[lb.RoundRobbinCount%len(lb.servers)]
	for !server.isAlive(){
		lb.RoundRobbinCount++
		server = lb.servers[lb.RoundRobbinCount%len(lb.servers)]
	}
	lb.RoundRobbinCount++
	return server
}

func (lb* LoadBalancer) serveProxy(r http.ResponseWriter,w *http.Request){
	targetServer := lb.getNextAvailableServer()
	fmt.Println("forwarding request to address %s",targetServer.address())
	targetServer.Serve(r,w)
}

func main(){
	servers:= []Server{
		newSimpleServer("https://www.youtube.com"),
		newSimpleServer("https://www.google.com"),
		newSimpleServer("https://www.instagram.com"),
	}
	lb := newLoadBalancer("8000",servers)
	handleRedirect :=func(r http.ResponseWriter,w *http.Request){
		lb.serveProxy(r,w)
	}
	http.HandleFunc("/",handleRedirect)

	fmt.Printf("serving requests at localhost:%s\n",lb.port)
	http.ListenAndServe(":"+lb.port,nil)
}