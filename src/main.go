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

	printMsg("STARTING")
	err := execCmd(startCommand)
	if err != nil {
		log.Fatal(err)
	}
	handleChildProcess()
	printMsg("START FINISHED")

	c := make(chan os.Signal, 1)
	signal.Notify(c)
	for {
		select {
		case sig := <-c:
			switch sig {
			case syscall.SIGHUP:
				fallthrough
			case syscall.SIGINT:
				fallthrough
			case syscall.SIGQUIT:
				fallthrough
			case syscall.SIGTERM:
				printMsg(fmt.Sprint("GOT SIGNAL [", sig.String(), "] STOPPING"))
				go handleStop(stopCommand)
			case syscall.SIGCHLD:
				go handleChildProcess()
			}
		}
	}
}

func execCmd(cmdStr string) error {
	cmd := exec.Command("/bin/bash", "-c", cmdStr)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	go pipeReader(bufio.NewReader(stdout))
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	go pipeReader(bufio.NewReader(stderr))
	err = cmd.Start()
	if err != nil {
		return err
	}
	cmd.Wait()
	return nil
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
	err := execCmd(stopCommand)
	if err != nil {
		log.Fatal(err)
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
