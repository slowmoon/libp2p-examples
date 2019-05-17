package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	inet "github.com/libp2p/go-libp2p-net"
	protocol "github.com/libp2p/go-libp2p-protocol"
	"github.com/multiformats/go-multiaddr"
	"os"
)

func  handleStream(stream inet.Stream)  {
      fmt.Println("got a new steam....")
      rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
      go readData(rw)
      go writeData(rw)
}


func readData(stream *bufio.ReadWriter)  {
    for {
         line, err := stream.ReadString('\n')
         if err!= nil{
         	fmt.Println("error reading data")
         	panic(err)
		 }
         if line == "" || len(line)==0 {
			 return
		 }

         if line!= "\n"{
         	fmt.Printf("\x1b[32m%s\x1b[0m>", line)
		 }
	}
}

func writeData(stream *bufio.ReadWriter)  {
   reader := bufio.NewReader(os.Stdin)
   for {
      fmt.Print(">")
      line, err := reader.ReadString('\n')
      if err!= nil{
      	fmt.Printf("Err reading from the stream %s", err.Error())
      	panic(err)
	  }
      _, err  = stream.WriteString(fmt.Sprintf("%s\n", line))
      if err!=nil {
      	fmt.Printf("Err writing data to stream %s", err.Error())
      	panic(err)
	  }
   }
}
func main() {

	help := flag.Bool("help", false, "whether need help")
	config := ParseFlags()

	if *help {
		fmt.Println("this is for help")
	}

	fmt.Printf("server startup with listen host %s and port %d\n", config.listenHost, config.listenPort)

	pri, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)

	if err!=nil {
		panic(err)
	}
	ctx := context.Background()

	addr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ipv4/%s/tcp/%d", config.listenHost, config.listenPort))

	host, err := libp2p.New(
		ctx,
		libp2p.ListenAddrs(addr),
		libp2p.Identity(pri),
	)

	if err!= nil{
		panic(err)
	}

	host.SetStreamHandler(protocol.ID(config.protocolId), handleStream)

	fmt.Printf("\n[*] Your Multiaddress is :/ipv4/%s/tcp/%d/p2p/%s\n",config.listenHost, config.listenPort, host.ID().Pretty())

	notifeeChan, err := NewDiscoveryNotifee(ctx, host, config.RendezvousString)

	if err!= nil {
		panic(err)
	}

	peer :=  <- notifeeChan

	fmt.Printf("find peer %v connection\n",peer)

	if err := host.Connect(ctx, peer);err!=nil {
		fmt.Printf("host connection fail %s\n",err.Error() )
	}

	hs ,err := host.NewStream(ctx, peer.ID, protocol.ID(config.protocolId))

	if err!= nil {

		fmt.Printf("new stream error %s\n", err.Error())

	}else {
		rw := bufio.NewReadWriter(bufio.NewReader(hs), bufio.NewWriter(hs))
		go readData(rw)
		go writeData(rw)
	}

	select {}
}
