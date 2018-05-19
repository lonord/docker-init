package main

/*
#include <stdlib.h>
#include <unistd.h>
#include <sys/wait.h>
#include <errno.h>

int checkpid() {
	return waitpid(-1, NULL, WNOHANG);
}
*/
import "C"

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {
	argc := len(os.Args)
	if argc != 3 {
		printUsage()
		return
	}
	startCommand := os.Args[1]
	stopCommand := os.Args[2]

	startCommandChan := make(chan int)
	go execStart(startCommandChan, startCommand)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan)
	for {
		select {
		case sig := <-signalChan:
			switch sig {
			case syscall.SIGHUP:
				fallthrough
			case syscall.SIGINT:
				fallthrough
			case syscall.SIGQUIT:
				fallthrough
			case syscall.SIGTERM:
				printMsg(fmt.Sprint("GOT SIGNAL [", sig.String(), "] STOPPING"))
				go killStartCommand(startCommandChan)
				go handleStop(stopCommand)
			case syscall.SIGCHLD:
				go handleChildProcess()
			}
		}
	}
}

func killStartCommand(startCommandChan chan int) {
	startCommandChan <- 1
}

func execStart(startCommandChan chan int, startCommand string) {
	printMsg("STARTING")
	cmd, err := execCmd(startCommand)
	if err != nil {
		log.Fatal(err)
	}
	select {
	case <-startCommandChan:
		if cmd != nil {
			cmd.Process.Kill()
		}
	}
}

func execCmd(cmdStr string) (*exec.Cmd, error) {
	cmd := exec.Command("/bin/bash", "-c", cmdStr)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	go pipeReader(bufio.NewReader(stdout))
	go pipeReader(bufio.NewReader(stderr))
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func pipeReader(reader *bufio.Reader) {
	for {
		line, err := reader.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		fmt.Printf(line)
	}
}

func handleChildProcess() {
	for {
		result := C.checkpid()
		if result <= 0 {
			break
		}
		printMsg(fmt.Sprint("REAP ZOMBIE CHILD [", result, "]"))
	}
}

func handleStop(stopCommand string) {
	cmd, err := execCmd(stopCommand)
	if err != nil {
		log.Fatal(err)
	}
	if cmd != nil {
		cmd.Wait()
	}
	handleChildProcess()
	printMsg("STOP FINISHED")
	os.Exit(0)
}

func printUsage() {
	fmt.Println("Usage: docker-init <start command> <stop command>")
}

func printMsg(msg string) {
	fmt.Println("\x1B[32m>> ", msg, " <<\x1B[0m")
}
