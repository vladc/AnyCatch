package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"log"
	"net"
	"os"
)

func main() {
	var anycastIP *string = flag.String("a", "1.2.3.4", "anycastip")
	var targetIP *string = flag.String("t", "1.2.3.4", "targetIP")

	flag.Usage = func() {
		log.Printf("usage: %s [ -i interface ] [ -a anycastip ] [ -s snaplen ] [ -X ] [ expression ]\n", os.Args[0])
		os.Exit(1)
	}

	flag.Parse()

	if *anycastIP == "1.2.3.4" || *targetIP == "1.2.3.4" {
		flag.Usage()
	}

	raddr := &net.IPAddr{IP: net.ParseIP(*targetIP).To4()}
	laddr := &net.IPAddr{IP: net.ParseIP(*anycastIP)}

	con, err := net.DialIP("ip4:1", laddr, raddr)
	if err != nil {
		log.Fatalf("unable to make raw socket to dial out from, err was %s", err.Error())
	}

	// Now to hand craft a ICMP echo request packet!

	// pkt := p.Payload
	// icmp := new(Icmphdr)
	// icmp.Type = pkt[0]
	// icmp.Code = pkt[1]
	// icmp.Checksum = binary.BigEndian.Uint16(pkt[2:4])
	// icmp.Id = binary.BigEndian.Uint16(pkt[4:6])
	// icmp.Seq = binary.BigEndian.Uint16(pkt[6:8])
	// p.Payload = pkt[8:]
	// p.Headers = append(p.Headers, icmp)
	// return icmp

	payload := []byte("ANYCATCH")
	packet := make([]byte, 7+len(payload)+1) // 7 for the packet itself, 8 for the "ANYCATCH" string

	packet[0] = 8 // Type, in this case a echo request
	packet[1] = 0 // Code, in this case there is no sub code for this packet type

	packet[2] = 0 // checksum of packet, we will do this later when we are done
	packet[3] = 0 // checksum of packet, we will do this later when we are done

	packet[4] = 69 // ICMP ID of the request, in this case I am just filling this in
	packet[5] = 69 // ICMP ID of the request

	packet[6] = 69 // ICMP Seq of the request, in this case I am just filling this in
	packet[7] = 69 // ICMP Seq of the request

	for i := 0; i < len(payload); i++ {
		packet[8+i] = payload[i]
	}
	csum, _ := getChecksum(packet)
	packet[2] = byte(csum >> 8)
	packet[3] = byte(csum & 255)

	con.Write(packet)
	log.Printf("Done.")

}

func getChecksum(data []byte) (uint16, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, data)
	if err != nil {
		return 0, err
	}
	arr := append(buf.Bytes(), data...)

	var sum uint32
	countTo := (len(arr) / 2) * 2

	// Sum as if we were iterating over uint16's
	for i := 0; i < countTo; i += 2 {
		p1 := (uint32)(arr[i+1]) * 256
		p2 := (uint32)(arr[i])
		sum += p1 + p2
	}

	// Potentially sum the last byte
	if countTo < len(arr) {
		sum += (uint32)(arr[len(arr)-1])
	}

	// Fold into 16 bits.
	sum = (sum >> 16) + (sum & 0xFFFF)
	sum = sum + (sum >> 16)

	// Take the 1's complement, and swap bytes.
	answer := ^((uint16)(sum & 0xFFFF))
	answer = (answer >> 8) | ((answer << 8) & 0xFF00)

	return answer, nil
}
