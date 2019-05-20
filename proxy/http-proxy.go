package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
	"io"
	"log"
	"net/http"
	"strings"
)

const  protocol  =  "/http-example/1.0.0"

func makeHost(port int)  (host host.Host, err error)  {
    ctx := context.Background()

	host, err  = libp2p.New(ctx,
			libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port)),
			)
	return
}

type ProxyService struct {
    host host.Host
    dest  peer.ID
    proxyAddr multiaddr.Multiaddr
}

func NewProxyService(host host.Host,proxyAddr multiaddr.Multiaddr, dest peer.ID) *ProxyService {
	host.SetStreamHandler(protocol, streamhandler)

	fmt.Println("proxy server is ready")
	fmt.Println("libp2p-peer addresses")

	for _, a := range host.Addrs() {
		fmt.Printf("%s/ipfs/%s\n", a, peer.IDB58Encode(host.ID()))
	}

	return &ProxyService{
		host: host,
		dest: dest,
	    proxyAddr: proxyAddr,
	}
}

func (s *ProxyService)Serve()  {
	  _, ps , _ := manet.DialArgs(s.proxyAddr)

	  fmt.Printf("proxy listen on : %s \n", ps )
	  if s.dest!= "" {
	     http.ListenAndServe(ps, s)
	  }
}

func (s *ProxyService)ServeHTTP(w http.ResponseWriter, r *http.Request)  {

	  fmt.Printf("proxying request for %s to peer %s\n", r.URL, s.dest.Pretty())

	  stream, err := s.host.NewStream(context.Background(), s.dest, protocol)
	  if err!= nil {
	  	fmt.Printf("error %s\n",  err.Error())
		  return
	  }
	  defer stream.Close()

	  err = r.Write(stream)
	  if err!= nil {
	  	   stream.Reset()
	  	   log.Println(err)
	  	   http.Error(w, err.Error(), http.StatusServiceUnavailable)
		  return
	  }

	 reader :=  bufio.NewReader(stream)
	 resp, err := http.ReadResponse(reader, r)

	 if err!= nil {
	 	stream.Reset()
	 	log.Println(err)
	 	http.Error(w, err.Error(), http.StatusServiceUnavailable)
	 	return
	 }
	 defer resp.Body.Close()
	 for k, v := range resp.Header {
	 	 for _, vv := range v {
	 	 	w.Header().Add(k, vv)
		 }
	 }

	 w.WriteHeader(resp.StatusCode)
	 io.Copy(w, resp.Body)

}


func streamhandler(stream net.Stream){
	defer stream.Close()

	bufreader := bufio.NewReader(stream)
	request, err := http.ReadRequest(bufreader)

	if err!=nil {
		stream.Reset()
		log.Println("reset stream...")
		return
	}
	defer request.Body.Close()

   //why reset
   request.URL.Host = "http"
   hp := strings.Split(request.Host, ":")
   if len(hp) >1 && hp[1] == "443" {
	   request.URL.Scheme = "https"
   }else {
	   request.URL.Scheme = "http"
   }

   outreq := new(http.Request)
   *outreq = *request

   resp, err := http.DefaultTransport.RoundTrip(outreq)
   if err != nil {
   	 stream.Reset()
   	 log.Println("send request fail")
	   return
   }
   defer resp.Body.Close()

   resp.Write(stream)

}

func addAddrToPeerStore(n host.Host, addr string) peer.ID  {

	muladdrs , err := multiaddr.NewMultiaddr(addr)

	if err!= nil {
		log.Panic(err)
	}

	pid, err :=  muladdrs.ValueForProtocol(multiaddr.P_IPFS)

	if err!= nil {
		log.Panic(err)
	}

	peerId, err := peer.IDB58Decode(pid)

	if err!= nil {
		log.Panic(err)
	}

	naddr, _ := multiaddr.NewMultiaddr(
	    fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerId)),
	)
	n.Peerstore().AddAddr(peerId, naddr,peerstore.PermanentAddrTTL)
	return peerId
}

func  main()  {

	 destPeer := flag.String("d", "", "special the port ")
	 port := flag.Int("port", 9000, "proxy port")
	 p2pport := flag.Int("1", 12000, "libp2p listen port")

	 if *destPeer!= "" {

	 	host, _ := makeHost(*port +1)
	 	peerId := addAddrToPeerStore(host, *destPeer)
	 	proxyaddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", *port))
	 	if err != nil {
	 		log.Panic(err)
		}
	 	proxy := NewProxyService(host, proxyaddr, peerId)
	 	proxy.Serve()
	 }else {
	    host , _ := makeHost(*p2pport)
	     _ = NewProxyService(host, nil, "")
		 select {}
	 }
}

