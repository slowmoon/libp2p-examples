package  main

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
)

func main()  {

	 ctx , cancelfunc:= context.WithCancel(context.Background())
     defer  cancelfunc()

	 host , err := libp2p.New(ctx)
	 if err!= nil {
	 	panic(err)
	 }

	 fmt.Printf("random create host addr %s\n", host.ID().Pretty())
	 //pri, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	 pri, _, err := crypto.GenerateECDSAKeyPair(rand.Reader)

	 if err!= nil {
	 	panic(err)
	 }

	 p2p , err := libp2p.New(
	    ctx,
	    libp2p.Identity(pri),
	    libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%d", "0.0.0.0", 8888)),
	 	)
	 if err!= nil{
	 	panic(err)
	 }

	fmt.Printf("remote peer id %s\n",p2p.ID()  )
}




