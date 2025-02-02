package helpers

import (
	"bufio"
	"context"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/TrueBlocks/trueblocks-key/test/dbtest"
)

var cancelSam context.CancelFunc

// StartSam starts AWS SAM local lambda as a separate process. This allows
// the tests to call lambda functions locally. cancelSam must be called
// to kill SAM process.
func StartSam() (cancelSam context.CancelFunc, wait func() error) {
	// Start AWS SAM local lambda
	awsProfile := os.Getenv("KY_TEST_AWS_PROFILE")
	if awsProfile == "" {
		awsProfile = "default"
	}
	dockerNetwork := dbtest.ContainerNetwork()
	if err := startDockerNetwork(dockerNetwork); err != nil {
		panic(err)
	}

	var samCtx context.Context
	samCtx, cancelSam = context.WithCancel(context.Background())
	samCmd := exec.CommandContext(
		samCtx,
		"sam",
		"local",
		"start-lambda",
		"--profile",
		awsProfile,
		"--docker-network",
		dockerNetwork,
		"--skip-pull-image",
	)
	samReady := make(chan bool)

	_, sourceFileName, _, ok := runtime.Caller(0)
	if !ok {
		panic("cannot read workdir")
	}
	samCmd.Dir = path.Join(filepath.Dir(sourceFileName), "../../../")
	log.Println("PWD is", samCmd.Dir)

	// Envs
	// Copy variables from the environment in which `go test` runs, so that
	// Docker can be found
	samCmd.Env = samCmd.Environ()
	dbEnvs := dbtest.ConnectionEnvs()
	// Add database connection envs
	for key, value := range dbEnvs {
		envVar := key + "=" + value
		log.Println(envVar)
		samCmd.Env = append(samCmd.Env, envVar)
	}

	go func() {
		log.Println("Waiting for SAM to start accepting connections...")
		stderr, err := samCmd.StderrPipe()
		if err != nil {
			panic(err)
		}
		if err := samCmd.Start(); err != nil {
			panic(err)
		}
		scanner := bufio.NewScanner(stderr)
		samRunning := false
		for scanner.Scan() {
			log.Println("output:", scanner.Text())
			if !samRunning && strings.Contains(scanner.Text(), "Running") {
				samReady <- true
				close(samReady)
				samRunning = true
			}
		}
	}()

	_ = <-samReady
	log.Println("SAM seems ready")
	return cancelSam, samCmd.Wait
}

func KillSamOnPanic() {
	if r := recover(); r != nil {
		cancelSam()
		panic(r)
	}
}

func startDockerNetwork(network string) error {
	cmd := exec.Command("docker", "network", "create", network)
	output, err := cmd.Output()
	log.Println("docker output:", string(output))
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok && strings.Contains(string(exitErr.Stderr), "already exists") {
			return nil
		}
		return err
	}

	return nil
}
