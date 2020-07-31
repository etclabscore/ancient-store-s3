package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

var (
	ipcPath = os.TempDir() + "ancient.ipc"
)

func runMain(testName string) {
	os.Setenv("AWS_REGION", "us-west-1")
	os.Setenv("AWS_PROFILE", "developers-s3")

	os.Args = append([]string{"./ancient-store-s3", "--bucket", fmt.Sprintf("etclabs-integration-test-%s", testName), "--loglevel", "3", "--ipcpath", ipcPath})
	main()

}

func getTestCases() []string {
	testCmd := exec.Command("go", "test", "github.com/ethereum/go-ethereum/core", "-list", "_RemoteFreezer")
	testCmd.Stderr = os.Stderr
	output, err := testCmd.Output()
	if err != nil {
		panic(err)
	}
	cases := strings.Split(string(output), "\n")
	parsedCases := []string{}
	for _, c := range cases {
		if strings.HasPrefix(c, "Test") {
			parsedCases = append(parsedCases, c)
		}
	}
	return parsedCases
}

func runTestCase(testCase string) {

	go func() {
		runMain(testCase)
	}()
	fmt.Println("TESTCASE===================:", testCase)
	testCmd := exec.Command("go", "test", "github.com/ethereum/go-ethereum/core", "-count=1", "-v", "-run", testCase)
	testCmd.Env = os.Environ()
	testCmd.Env = append(os.Environ(), fmt.Sprintf("GETH_ANCIENT_RPC=%s", ipcPath))
	testCmd.Stderr = os.Stderr
	testCmd.Stdout = os.Stdout
	err := testCmd.Run()
	if err != nil {
		panic(err)
	}
}

func TestIntegration(t *testing.T) {

	testCases := getTestCases()
	for _, testCase := range testCases {
		runTestCase(testCase)
	}
}
