package  main

import "flag"

type Config struct {
    RendezvousString string
    protocolId string
    listenHost string
    listenPort int
}


func ParseFlags()*Config {

     config := Config{}
     flag.StringVar(&config.RendezvousString, "rendezvous", "meetme", "unique the node of you address")
     flag.StringVar(&config.protocolId, "pid", "/chat/1.1.0", "set a protocol header for the stream")
     flag.StringVar(&config.listenHost, "host", "0.0.0.0", "special the host")
     flag.IntVar(&config.listenPort, "port", 8001, "node listen port")

     flag.Parse()
     return  &config
}


