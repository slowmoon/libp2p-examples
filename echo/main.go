package  main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/multiformats/go-multiaddr"
	"io"
	"io/ioutil"
	"log"
	mrand "math/rand"
)

func  makehost(listenport int , insecure bool, randseed int64)(host.Host, error) {

	  var r io.Reader
      if randseed == 0 {
         r = rand.Reader
	  }else {
	  	r =  mrand.New(mrand.NewSource(randseed))
	  }
	  pri, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	  if err != nil {
	  	panic(err)
	  }

	  options := []libp2p.Option{
	  	libp2p.Identity(pri),
	  	libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenport )),
	  	libp2p.DisableRelay(),
	  }

	  if !insecure {
	  	options = append(options, libp2p.NoSecurity)
	  }

	  host, err := libp2p.New(context.Background() , options...)

	  if err != nil {
	  	panic(err)
	  }

	  fmt.Println("host addr is :", host.ID())

	  hostaddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ipfs/%s", host.ID().Pretty()))

	  fullAddr := host.Addrs()[0].Encapsulate(hostaddr)

	  log.Printf("i am %s\n", fullAddr.String())

	  if insecure {
	  	  log.Printf("now run ipfs with insecure, localport %d fulladdr %s\n", listenport+1, fullAddr.String())
	  }else {
		  log.Printf("now run ipfs , localport %d fulladdr %s\n", listenport, fullAddr.String())
	  }

	return host, nil

}

func  echo(s net.Stream) error {
     log.Println("got a new stream!")
     reader := bufio.NewReader(s)
     res, err := reader.ReadString('\n')
     if err != nil {
     	log.Println("recev error", err.Error())
		 return err
	 }
     res = fmt.Sprintf("read %s\n", res)
     fmt.Printf("%s\n", res)
     _, err = s.Write([]byte(res))
     if err != nil {
		 log.Println("recev error", err.Error())
	 }

     return err
}



func main()  {

    insecure := flag.Bool("secure", true, "use no encrypt service")
    listenf := flag.Int("l", 8080, "listen the port")
    target := flag.String("d", "", "target address")
    seed := flag.Int64("s", 0, "crypto rand seed")

    flag.Parse()

    if *listenf ==0{
       log.Panic("please special a port for listen")
	}

   host ,err :=  makehost(*listenf, *insecure, *seed)
   if err != nil {
   	log.Panic("host listen error\n",err)
   }

   host.SetStreamHandler("/echo/1.0.0", func(s net.Stream) {
	    if err= echo(s);err!=nil {
	    	s.Reset()
		}else{
			s.Close()
		}
   })

   if *target== ""{
	   select {
	   }
	   log.Printf("target listen here...\n")
	   return
   }

   addr, err := multiaddr.NewMultiaddr(*target)
   if err != nil {
   	   log.Panic("fail to parse multiaddr ", err.Error())
   }

   pid, err := addr.ValueForProtocol(multiaddr.P_IPFS)

   if err != nil {
   	  log.Panic(err)
   }

   peerId, err := peer.IDB58Decode(pid)
   if err != nil {
   	  log.Panic(err)
   }

   ipfsaddr , err := multiaddr.NewMultiaddr(fmt.Sprintf("/ipfs/%s", pid))

   peeraddr := addr.Decapsulate(ipfsaddr)
   fmt.Printf("peer listen addr is %s\n", peeraddr.String())
   //add into the peerstore , let p2p know where to transfer
   host.Peerstore().AddAddr(peerId, peeraddr, peerstore.PermanentAddrTTL)

   st, err:= host.NewStream(context.Background(), peerId, "/echo/1.0.0")
   if err != nil {
		log.Panic(err)
   }

   st.Write([]byte("hello world!\n"))

   con, err := ioutil.ReadAll(st)
   if err != nil{
   	 log.Panic(err)
   }
   fmt.Printf("reply [%s]", con)
}
