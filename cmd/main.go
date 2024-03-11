package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

	"project.com/pkg/elevator"
	"project.com/pkg/infobank"
	"project.com/pkg/network"
)

const (
	heartbeatSleep_ms = 500
	deadLine_ms       = 5 * heartbeatSleep_ms
	numberOfFloors    = 4
	bufSize           = 50
)

func main() {
	backupProcess()
}

func startBackupProcess(port string) {
	exec.Command("gnome-terminal", "--", "go", "run", "main.go", port).Start()
}

func primaryProcess(lastID string, udpSendAddr string) {

	sendUDPAddr, err := net.ResolveUDPAddr("udp", udpSendAddr)
	if err != nil {
		fmt.Println(err)
		return
	}

	conn, err := net.DialUDP("udp", nil, sendUDPAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	//"Main logic"
	ID, elevStatusUpdate_ch := prepareSystem()
	elevator.ElevatorInit(elevStatusUpdate_ch, lastID, ID)

	for {
		msg := ID
		_, err := conn.Write([]byte(msg))
		if err != nil {
			fmt.Println("Primary failed to send heartbeat:", err)
			return
		}
		time.Sleep(heartbeatSleep_ms * time.Millisecond)
	}
}

func backupProcess() {
	port := os.Args[len(os.Args)-1] //Port is always last element of command line input
	udpReceiveAddr := ":" + port
	udpSendAddr := "255.255.255.255" + udpReceiveAddr

	receiveUDPAddr, err := net.ResolveUDPAddr("udp", udpReceiveAddr)
	if err != nil {
		fmt.Println(err)
		return
	}

	conn, err := net.ListenUDP("udp", receiveUDPAddr)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer conn.Close()

	elevator.Init("localhost:"+port, numberOfFloors)

	buffer := make([]byte, 1024)
	var lastID string
	for {
		conn.SetReadDeadline(time.Now().Add(deadLine_ms * time.Millisecond))
		n, _, err := conn.ReadFromUDP(buffer)

		if err == nil {
			lastID = string(buffer[:n])
		} else {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				conn.Close()
				startBackupProcess(port)
				primaryProcess(lastID, udpSendAddr)
				return
			} else {
				fmt.Println("Error reading from UDP:", err)
				return
			}
		}
	}
}

func prepareSystem() (ID string, elevStatusUpdate_ch chan elevator.Elevator) {
	elevStatusUpdate_ch = make(chan elevator.Elevator, bufSize)
	networkUpdateTx_ch := make(chan network.Msg, bufSize)
	networkUpdateRx_ch := make(chan network.Msg, bufSize)
	peerUpdate_ch := make(chan network.PeerUpdate, bufSize)

	go elevator.FSM(elevStatusUpdate_ch)
	go infobank.Infobank_FSM(elevStatusUpdate_ch, networkUpdateTx_ch, networkUpdateRx_ch, peerUpdate_ch)
	go network.Network_fsm(networkUpdateTx_ch, networkUpdateRx_ch, peerUpdate_ch)
	ID, _ = network.LocalIP()

	return ID, elevStatusUpdate_ch
}
