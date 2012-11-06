package main

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

func main() {

	// load certificates and setup config
	cert, err := tls.LoadX509KeyPair("./apns-prod-cert.pem", "./apns-prod-key.pem")
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
	conf := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// connect to the APNS and wrap socket to tls client
	conn, err := net.Dial("tcp", "gateway.sandbox.push.apple.com:2195")
	if err != nil {
		fmt.Printf("tcp error: %s\n", err)
		os.Exit(1)
	}
	tlsconn := tls.Client(conn, conf)

	// Force handshake to verify successful authorization.
	// Handshake is handled otherwise automatically on first
	// Read/Write attempt
	err = tlsconn.Handshake()
	if err != nil {
		fmt.Printf("tls error: %s\n", err)
		os.Exit(1)
	}
	// informational debugging stuff
	state := tlsconn.ConnectionState()
	fmt.Printf("conn state %v\n", state)

	// prepare binary payload from JSON structure
	payload := make(map[string]interface{})
	payload["aps"] = map[string]string{"alert": "Hello Push"}
	bpayload, err := json.Marshal(payload)

	// decode hexadecimal push device token to binary byte array
	btoken, _ := hex.DecodeString("这里写Token")

	// build the actual pdu
	buffer := bytes.NewBuffer([]byte{})
	// command
	binary.Write(buffer, binary.BigEndian, uint8(1))

	// transaction id, optional
	binary.Write(buffer, binary.BigEndian, uint32(1))

	// expiration time, 1 hour
	binary.Write(buffer, binary.BigEndian, uint32(time.Now().Unix()+60*60))

	// push device token
	binary.Write(buffer, binary.BigEndian, uint16(len(btoken)))
	binary.Write(buffer, binary.BigEndian, btoken)

	// push payload
	binary.Write(buffer, binary.BigEndian, uint16(len(bpayload)))
	binary.Write(buffer, binary.BigEndian, bpayload)
	pdu := buffer.Bytes()

	// write pdu
	_, err = tlsconn.Write(pdu)
	if err != nil {
		fmt.Printf("write error: %s\n", err)
		os.Exit(1)
	}

	// wait for 5 seconds error pdu from the socket
	// tlsconn.SetReadTimeout(5 * 1E9)

	readb := [6]byte{}
	n, err := tlsconn.Read(readb[:])
	if n > 0 {
		fmt.Printf("received: %s\n", hex.EncodeToString(readb[:n]))
	}

	tlsconn.Close()
}
