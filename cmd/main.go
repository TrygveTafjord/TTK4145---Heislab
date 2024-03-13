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
	"project.com/pkg/initialize"
)

const (
	heartbeatSleep = 500
)

func startBackupProcess(port string) {
	fmt.Print("I get here")
	exec.Command("gnome-terminal", "--", "go", "run", "main.go", port).Run()
}

func primaryProcess(lastID string, port string, udpSendAddr string) {
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

	const BUFFER_SIZE = 50

	requestUpdate_ch := make(chan [elevator.N_FLOORS][elevator.N_BUTTONS]bool, BUFFER_SIZE)
	clearRequest_ch := make(chan [elevator.N_FLOORS][elevator.N_BUTTONS]bool, BUFFER_SIZE)
	stateUpdate_ch := make(chan elevator.State, BUFFER_SIZE)
	lightsUpdate_ch := make(chan [elevator.N_FLOORS][elevator.N_BUTTONS]bool, BUFFER_SIZE)
	obstruction_ch := make(chan bool, BUFFER_SIZE)

	elevInitFSM_ch := make(chan elevator.Elevator, 50)
	elevInitInfobank_ch := make(chan infobank.ElevatorInfo)
	networkUpdateTx_ch := make(chan network.Msg, 50)
	networkUpdateRx_ch := make(chan network.Msg, 50)
	peerUpdate_ch := make(chan network.PeerUpdate, 50)

	go elevator.FSM(requestUpdate_ch, clearRequest_ch, stateUpdate_ch, lightsUpdate_ch, elevInitFSM_ch)
	go infobank.Infobank(requestUpdate_ch, clearRequest_ch, stateUpdate_ch, lightsUpdate_ch, networkUpdateTx_ch, networkUpdateRx_ch, peerUpdate_ch, obstruction_ch)
	//go network.Network_fsm(networkUpdateTx_ch, networkUpdateRx_ch, peerUpdate_ch)
	ID, err := network.LocalIP()

	initialize.ElevatorInit(elevInitInfobank_ch, elevInitFSM_ch, lastID, ID)

	for {
		msg := ID
		_, err := conn.Write([]byte(msg))
		if err != nil {
			fmt.Println("Primary failed to send heartbeat:", err)
			return
		}
		time.Sleep(heartbeatSleep * time.Millisecond)
	}

}

func backupProcess() {
	fmt.Printf("---------BACKUP PHASE---------\n")

	args := os.Args
	fmt.Println(args)
	port := args[1]
	fmt.Printf("PORT: %v", port)
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

	elevator.Init("localhost:"+port, 4)

	if err != nil {
		fmt.Printf("could not get IP")
	}

	var lastID string

	for {
		buffer := make([]byte, 1024)
		conn.SetReadDeadline(time.Now().Add(heartbeatSleep * 5 * time.Millisecond))
		n, _, err := conn.ReadFromUDP(buffer)

		if err == nil {
			lastID = string(buffer[:n])
		} else {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				conn.Close()
				startBackupProcess(port)
				primaryProcess(lastID, port, udpSendAddr)
				return
			} else {
				fmt.Println("Error reading from UDP:", err)
				return
			}
		}
	}
}

func main() {
	backupProcess()
}
