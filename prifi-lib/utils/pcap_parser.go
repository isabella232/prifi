package utils

import (
	"bufio"
	"encoding/binary"
	"errors"
	"github.com/Lukasa/gopcap"
	"go.dedis.ch/onet/v3/log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

const pattern uint16 = uint16(21845) //0101010101010101
const metaMessageLength int = 15     // 2bytes pattern + 4bytes ID + 8bytes timeStamp + 1 bit fragmentation

// Packet is an ID(Packet number), TimeSent in microsecond, and some Data
type Packet struct {
	ID                        uint32
	MsSinceBeginningOfCapture uint64 //milliseconds since beginning of capture
	Header                    []byte
	RealLength                int
}

// Parses a .pcap file, and returns all valid packets. A packet is (ID, TimeSent [milliseconds], Data)
func ParsePCAP(path string, maxPayloadLength int, clientID uint16) ([]Packet, error) {
	pcapfile, err := os.Open(path)
	if err != nil {
		return nil, errors.New("Cannot open" + path + "error is" + err.Error())
	}
	defer pcapfile.Close()

	parsed, err := gopcap.Parse(pcapfile)
	if err != nil {
		return nil, errors.New("Cannot parse" + path + "error is" + err.Error())
	}

	out := make([]Packet, 0)

	if len(parsed.Packets) == 0 {
		return out, nil
	}

	time0 := parsed.Packets[0].Timestamp.Nanoseconds()

	// Adds a random number \in [0, 10] sec to all times
	rand.Seed(time.Now().UTC().UnixNano())
	random_offset := uint64(rand.Intn(10000)) // r is in ms

	for id, pkt := range parsed.Packets {

		t := uint64((pkt.Timestamp.Nanoseconds()-time0)/1000000) + random_offset
		remainingLen := int(pkt.IncludedLen)

		//maybe this packet is bigger than the payload size. Then, generate many packets
		for remainingLen > maxPayloadLength {
			p2 := Packet{
				ID:                        uint32(id),
				Header:                    metaBytes(maxPayloadLength, clientID, uint32(id), t, false),
				MsSinceBeginningOfCapture: t,
				RealLength:                maxPayloadLength,
			}
			out = append(out, p2)
			remainingLen -= maxPayloadLength
		}

		//add the last packet, that will trigger the relay pattern match
		if remainingLen < metaMessageLength {
			remainingLen = metaMessageLength
		}
		p := Packet{
			ID:                        uint32(id),
			Header:                    metaBytes(remainingLen, clientID, uint32(id), t, true),
			MsSinceBeginningOfCapture: t,
			RealLength:                remainingLen,
		}
		out = append(out, p)
	}

	return out, nil
}

// Parses a .pkts file (homemade format with [timestamp, bytes, npackets], and returns all valid packets. A packet is (ID, TimeSent [milliseconds], Data)
func ParsePKTS(path string, maxPayloadLength int, clientID uint16) ([]Packet, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, errors.New("Cannot open" + path + "error is" + err.Error())
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	out := make([]Packet, 0)

	packetID := uint32(0)
	for scanner.Scan() {
		line := strings.Replace(scanner.Text(), "\t", "", -1)
		parts := strings.Split(line, ",")

		time_ms_str := strings.TrimSpace(parts[0])
		parts2 := strings.Split(time_ms_str, ".")
		time_str := strings.TrimSpace(parts2[0])
		time_ms := strings.TrimSpace(parts2[1])

		time_ms_i, err := strconv.Atoi(time_ms)
		if err != nil {
			log.Lvl1("Can't convert", time_ms, "to int")
		}

		bytes_str := strings.TrimSpace(parts[1])

		layout := "15:04:05"
		time_parsed, err := time.Parse(layout, time_str)
		if err != nil {
			log.Lvl1("Can't convert", time_str, "to time", err)
		}
		packet_time_ms := uint64(time_parsed.Hour()*3600*1000) + uint64(time_parsed.Minute()*60*1000) + uint64(time_parsed.Second()*1000) + uint64(time_ms_i)

		bytes, err := strconv.Atoi(bytes_str)
		if err != nil {
			log.Lvl1("Can't convert", bytes_str, "to int")
		}
		remainingLen := bytes

		//maybe this packet is bigger than the payload size. Then, generate many packets
		for remainingLen > maxPayloadLength {
			p2 := Packet{
				ID:                        uint32(packetID),
				Header:                    metaBytes(maxPayloadLength, clientID, uint32(packetID), packet_time_ms, false),
				MsSinceBeginningOfCapture: packet_time_ms,
				RealLength:                maxPayloadLength,
			}
			out = append(out, p2)
			remainingLen -= maxPayloadLength
		}

		//add the last packet, that will trigger the relay pattern match
		if remainingLen < metaMessageLength {
			remainingLen = metaMessageLength
		}
		p := Packet{
			ID:                        uint32(packetID),
			Header:                    metaBytes(remainingLen, clientID, uint32(packetID), packet_time_ms, true),
			MsSinceBeginningOfCapture: packet_time_ms,
			RealLength:                remainingLen,
		}
		out = append(out, p)

		packetID++
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.New("Cannot read" + path + "error is" + err.Error())
	}
	return out, nil
}

func getPayloadOrRandom(pkt gopcap.Packet, clientID uint16, packetID uint32, msSinceBeginningOfCapture uint64) []byte {
	len := pkt.IncludedLen

	if true || pkt.Data == nil {
		return metaBytes(int(len), clientID, packetID, msSinceBeginningOfCapture, false)
	}

	return pkt.Data
}

func metaBytes(length int, clientID uint16, packetID uint32, timeSentInPcap uint64, isFinalPacket bool) []byte {
	// ignore length, have short messages
	if false && length < metaMessageLength {
		return recognizableBytes(length, packetID)
	}
	// out := make([]byte, length)
	out := make([]byte, 17)
	binary.BigEndian.PutUint16(out[0:2], pattern)
	binary.BigEndian.PutUint16(out[2:4], clientID)
	binary.BigEndian.PutUint32(out[4:8], packetID)
	binary.BigEndian.PutUint64(out[8:16], timeSentInPcap)
	out[16] = byte(0)
	if isFinalPacket {
		out[16] = byte(1)
	}
	return out
}

func recognizableBytes(length int, packetID uint32) []byte {
	if length == 0 {
		return make([]byte, 0)
	}
	pattern := make([]byte, 4)
	binary.BigEndian.PutUint32(pattern, packetID)

	pos := 0
	out := make([]byte, length)
	for pos < length {
		//copy from pos,
		copyLength := len(pattern)
		copyEndPos := pos + copyLength
		if copyEndPos > length {
			copyEndPos = length
			copyLength = copyEndPos - pos
		}
		copy(out[pos:copyEndPos], pattern[0:copyLength])
		pos = copyEndPos
	}

	return out
}

func randomBytes(len uint32) []byte {
	if len == uint32(0) {
		return make([]byte, 0)
	}
	out := make([]byte, len)
	written, err := rand.Read(out)
	if err == nil {
		log.Fatal("Could not generate a random packet of length", len, "error is", err)
	}
	if uint32(written) != len {
		log.Fatal("Could not generate a random packet of length", len, "only wrote", written)
	}
	return out
}
