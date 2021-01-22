package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

func execCommand(cmdAndArgs []string) (string, error) {
	var (
		stdOut bytes.Buffer
		stdErr bytes.Buffer
	)
	command := cmdAndArgs[0]
	arguments := cmdAndArgs[1:]
	cmd := exec.Command(command, arguments...)
	cmd.Stderr = &stdErr
	cmd.Stdout = &stdOut
	err := cmd.Start()
	if err != nil {
		return "", err
	}
	err = cmd.Wait()
	if err != nil {
		outStr := stdOut.String()
		if len(outStr) > 0 {
			fmt.Printf("STDOUT:\n%v\n", outStr)
		}
		errStr := stdErr.String()
		if len(errStr) > 0 {
			fmt.Printf("STDERR:\n%v\n", errStr)
		}
		return stdOut.String(), err
	}
	outStr := stdOut.String()
	return outStr, nil
}

func execCommandWithStdin(cmdAndArgs []string, stdIn *bytes.Buffer) (string, error) {
	var (
		stdOut bytes.Buffer
		stdErr bytes.Buffer
	)
	command := cmdAndArgs[0]
	arguments := cmdAndArgs[1:]
	cmd := exec.Command(command, arguments...)
	cmd.Stderr = &stdErr
	cmd.Stdout = &stdOut
	cmd.Stdin = stdIn
	err := cmd.Start()
	if err != nil {
		return "", err
	}
	err = cmd.Wait()
	if err != nil {
		outStr := stdOut.String()
		if len(outStr) > 0 {
			fmt.Printf("STDOUT:\n%v\n", outStr)
		}
		errStr := stdErr.String()
		if len(errStr) > 0 {
			fmt.Printf("STDERR:\n%v\n", errStr)
		}
		return stdOut.String(), err
	}
	outStr := stdOut.String()
	return outStr, nil
}
