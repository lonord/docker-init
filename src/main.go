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
	stopCommand := os.Args[1]

	err := execCmd(startCommand)
	if err != nil {
		log.Fatal(err)
	}
	handleChildProcess()
	fmt.Println("> START FINISHED")

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
				err := execCmd(stopCommand)
				if err != nil {
					log.Fatal(err)
				}
				handleChildProcess()
				fmt.Println("> STOP FINISHED")
				os.Exit(0)
			case syscall.SIGCHLD:
				handleChildProcess()
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
		if result := C.checkpid(); result <= 0 {
			break
		}
	}
}

func printUsage() {
	fmt.Println("Usage: docker-init <start command> <stop command>")
}
