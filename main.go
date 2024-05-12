package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

const (
	filePath = "test.txt"
)

func fileContents() string {
	f, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	contents, err := io.ReadAll(f)
	if err != nil {
		return ""
	}
	return string(contents)
}

func main() {
	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
	)
	if err != nil {
		panic(err)
	}

	dest := flag.String("d", "", "destination Peer address")
	flag.Parse()

	nodeInfo := peerstore.AddrInfo{
		ID:    node.ID(),
		Addrs: node.Addrs(),
	}
	addrs, err := peerstore.AddrInfoToP2pAddrs(&nodeInfo)
	if err != nil {
		panic(err)
	}
	fmt.Println("host node address:", addrs[0])

	if *dest == "" {
		myContents := fileContents()
		node.SetStreamHandler("/dfs/1.0.0", func(s network.Stream) {
			rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
			fmt.Printf("Remote peer addres: %+v\n", s.Conn().RemotePeer())
			nBytes, err := rw.WriteString(myContents)
			if err != nil {
				log.Printf("Couldn't write to stream: %s\n", err.Error())
				os.Exit(1)
			}
			fmt.Printf("Wrote %d bytes to the stream: ", nBytes)
		})

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
	} else {
		addr, err := multiaddr.NewMultiaddr(*dest)
		if err != nil {
			log.Printf("Cannot parse multi address of node: %s\n", err.Error())
			os.Exit(1)
		}
		peer, err := peerstore.AddrInfoFromP2pAddr(addr)
		if err != nil {
			log.Printf("Cannot parse multi address of node: %s\n", err.Error())
			os.Exit(1)
		}
		if err := node.Connect(context.Background(), *peer); err != nil {
			panic(err)
		}

		s, err := node.NewStream(context.Background(), peer.ID, "/dfs/1.0.0")
		if err != nil {
			log.Printf("Cannot start a new stream: %s\n", err.Error())
			os.Exit(1)
		}

		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
		bytes, err := rw.ReadString('\n')
		if err != nil {
			log.Printf("Cannot read from read writer.: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Println("Read the following bytes from stream: ", string(bytes))
	}
}
