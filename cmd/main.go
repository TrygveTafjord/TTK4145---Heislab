package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

	"project.com/pkg/diagnostics"
	"project.com/pkg/elevator"
	"project.com/pkg/infobank"
	"project.com/pkg/initialize"
	"project.com/pkg/network"
)

const (
	heartbeatSleep = 500
	BUFFER_SIZE = 50
	numFloors = 4
)

func startBackupProcess(port string) {
	exec.Command("gnome-terminal", "--", "go", "run", "main.go", port).Run()
}

func primaryProcess(sendAddr string) {

	sendUDPAddr, err := net.ResolveUDPAddr("udp", sendAddr)
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

	initFSM_ch := make(chan elevator.Elevator, BUFFER_SIZE)
	initInfobank_ch := make(chan infobank.ElevatorInfo, BUFFER_SIZE)
	initNetwork_ch := make(chan string, BUFFER_SIZE)

	requestUpdate_ch := make(chan [elevator.N_FLOORS][elevator.N_BUTTONS]bool, BUFFER_SIZE)
	clearRequest_ch := make(chan []elevator.ButtonEvent, BUFFER_SIZE)
	stateUpdate_ch := make(chan elevator.State, BUFFER_SIZE)
	lightsUpdate_ch := make(chan [elevator.N_FLOORS][elevator.N_BUTTONS]bool, BUFFER_SIZE)
	obstruction_ch := make(chan bool)

	requestInfobankToNetwork_ch := make(chan network.NewRequest, BUFFER_SIZE)
	requestNetworkToInfobank_ch := make(chan network.NewRequest, BUFFER_SIZE)
	obstructedInfobankToNetwork_ch := make(chan network.Obstructed, BUFFER_SIZE)
	obstructedNetworkToInfobank_ch := make(chan network.Obstructed, BUFFER_SIZE)
	stateInfobankToNetwork_ch := make(chan network.StateUpdate, BUFFER_SIZE)
	stateNetworkToInfobank_ch := make(chan network.StateUpdate, BUFFER_SIZE)
	clearedInfobankToNetwork_ch := make(chan network.RequestCleared, BUFFER_SIZE)
	clearedNetworkToInfobank_ch := make(chan network.RequestCleared, BUFFER_SIZE)
	sendRequestConfirmation_ch := make(chan network.Confirm, BUFFER_SIZE)
	recieveRequestConfirmation_ch := make(chan network.Confirm, BUFFER_SIZE)
	periodicInfobankToNetwork_ch := make(chan network.Periodic, BUFFER_SIZE)
	periodicNetworkToInfobank_ch := make(chan network.Periodic, BUFFER_SIZE)

	updateDiagnostics_ch := make(chan elevator.Elevator)
	obstructionDiagnoze_ch := make(chan bool)
	peerUpdate_ch := make(chan network.PeerUpdate, 50)

	go elevator.FSM(
		initFSM_ch,
		requestUpdate_ch,
		clearRequest_ch,
		stateUpdate_ch,
		lightsUpdate_ch,
		obstruction_ch,
		updateDiagnostics_ch,
		obstructionDiagnoze_ch)

	go infobank.Infobank(
		initInfobank_ch,
		requestUpdate_ch,
		clearRequest_ch,
		stateUpdate_ch,
		lightsUpdate_ch,
		obstruction_ch,
		requestInfobankToNetwork_ch,
		requestNetworkToInfobank_ch,
		sendRequestConfirmation_ch,
		recieveRequestConfirmation_ch,
		obstructedInfobankToNetwork_ch,
		obstructedNetworkToInfobank_ch,
		stateInfobankToNetwork_ch,
		stateNetworkToInfobank_ch,
		clearedInfobankToNetwork_ch,
		clearedNetworkToInfobank_ch,
		periodicInfobankToNetwork_ch,
		periodicNetworkToInfobank_ch,
		peerUpdate_ch)

	go network.Network(
		initNetwork_ch,
		requestNetworkToInfobank_ch,
		requestInfobankToNetwork_ch,
		recieveRequestConfirmation_ch,
		sendRequestConfirmation_ch,
		obstructedNetworkToInfobank_ch,
		obstructedInfobankToNetwork_ch,
		stateNetworkToInfobank_ch,
		stateInfobankToNetwork_ch,
		clearedNetworkToInfobank_ch,
		clearedInfobankToNetwork_ch,
		periodicInfobankToNetwork_ch,
		periodicNetworkToInfobank_ch,
		peerUpdate_ch)

	go diagnostics.Diagnostics(updateDiagnostics_ch, obstructionDiagnoze_ch)

	id, _ := network.LocalIP()

	initialize.ElevatorInit(initInfobank_ch, initFSM_ch, initNetwork_ch, id)

	for {
		msg := id
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
	receiveAddr := ":" + port
	udpSendAddr := "255.255.255.255" + receiveAddr

	receiveUDPAddr, err := net.ResolveUDPAddr("udp", receiveAddr)
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

	id, err := network.LocalIP()
	if err != nil {    
		fmt.Println(err)
		return
	}
	
	elevator.Init(id + ":15657", numFloors)

	for {
		buffer := make([]byte, 1024)
		conn.SetReadDeadline(time.Now().Add(heartbeatSleep * 10 * time.Millisecond))
		_, _, err := conn.ReadFromUDP(buffer)

		if err == nil {
		} else {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				conn.Close()
				startBackupProcess(port)
				primaryProcess(udpSendAddr)
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
