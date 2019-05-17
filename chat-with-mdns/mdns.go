package main

import (
    "context"
    host "github.com/libp2p/go-libp2p-host"
    peerstore "github.com/libp2p/go-libp2p-peerstore"
    "github.com/libp2p/go-libp2p/p2p/discovery"
    "time"
)

type discoveryNotifee struct {
    Peerchan   chan peerstore.PeerInfo
}

func (d *discoveryNotifee)HandlePeerFound(pi peerstore.PeerInfo)  {
     d.Peerchan <- pi
}

func NewDiscoveryNotifee(ctx context.Context,peerHost host.Host, rendezous string)(chan peerstore.PeerInfo, error)  {

     service , err:= discovery.NewMdnsService(ctx,  peerHost, time.Hour, rendezous)

     if err!=nil {
     	panic(err)
     }

     discovery:= discoveryNotifee{
         Peerchan: make(chan peerstore.PeerInfo),
     }

     service.RegisterNotifee(&discovery)
     return  discovery.Peerchan, nil
}
