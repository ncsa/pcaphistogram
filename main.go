package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
)

func histogram(r gopacket.PacketDataSource, lt layers.LinkType) (int, []uint64, error) {
	hist := make([]uint64, 256)
	var err error
	var packets int
	var packet gopacket.Packet
	ps := gopacket.NewPacketSource(r, lt)
	for {
		packet, err = ps.NextPacket()
		if packet == nil {
			break
		}
		packets++
		if app := packet.ApplicationLayer(); app != nil {
			for _, b := range app.LayerContents() {
				hist[b]++
			}
		}
		if el := packet.ErrorLayer(); el != nil {
			fmt.Fprintf(os.Stderr, "error decoding packet %d: %q\n", packets, el)
		}
	}
	if err == io.EOF {
		err = nil
	}
	return packets, hist, err
}

func getHistogram(filename string) (int, []uint64, error) {
	inf, _ := os.Open(filename)
	defer inf.Close()
	r, err := pcapgo.NewReader(inf)
	if err != nil {
		panic(err)
	}
	return histogram(r, r.LinkType())
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n %s hist infile.pcap\n %s plot infile.pcap out.png\n", os.Args[0], os.Args[0])
}

func runHist(args []string) int {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Missing filename\n")
		usage()
		return 1
	}
	packets, hist, err := getHistogram(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "%d packets processed\n", packets)
	for b, freq := range hist {
		fmt.Printf("%d\t%d\n", b, freq)
	}
	return 0
}

func runPlot(args []string) int {
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Missing filename\n")
		usage()
		return 1
	}
	packets, hist, err := getHistogram(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return 1
	}
	fmt.Printf("%d packets processed\n", packets)
	return 0
}

func main() {
	flag.Parse()

	if len(flag.Args()) <= 1 {
		usage()
		os.Exit(1)
	}

	mode := flag.Args()[0]
	if mode == "hist" {
		os.Exit(runHist(flag.Args()[1:]))
	} else if mode == "plot" {
		os.Exit(runPlot(flag.Args()[1:]))
	} else {
		usage()
		os.Exit(1)
	}
}
