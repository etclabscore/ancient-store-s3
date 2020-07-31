package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
)

func init() {
	if os.Getenv("AWS_REGION") == "" {
		log.Println("Setting default AWS_REGION credential")
		os.Setenv("AWS_REGION", "us-west-1")
	}
	if os.Getenv("AWS_PROFILE") == "" {
		log.Println("Setting default AWS_PROFILE credential")
		os.Setenv("AWS_PROFILE", "developers-s3")
	}
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

func runTestCase(testCase string, t *testing.T) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "freezer_s3_test")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	ipcPath := filepath.Join(tmpDir, "ancient.ipc")
	go func() {
		// https://docs.aws.amazon.com/AmazonS3/latest/dev/BucketRestrictions.html
		bucketName := fmt.Sprintf("ancientstore-%s", testCase)
		bucketName = strings.ReplaceAll(bucketName, "_", "")
		bucketName = strings.ToLower(bucketName)
		if len(bucketName) > 63 {
			bucketName = bucketName[:63]
		}
		if err := app.Run([]string{"ancient-store-s3", "--bucket", bucketName, "--loglevel", "3", "--ipcpath", ipcPath}); err != nil {
			fmt.Fprintln(os.Stderr, err)
			log.Printf("App exited erroring: %v", err)
			os.Exit(1)
		}
		fmt.Println("App exited 0")
	}()
	fmt.Println("TESTCASE===================:", testCase)
	testCmd := exec.Command("go", "test", "github.com/ethereum/go-ethereum/core", "-count=1", "-v", "-run", testCase)
	testCmd.Stderr = os.Stderr
	testCmd.Stdout = os.Stdout

	testCmd.Env = os.Environ()
	testCmd.Env = append(testCmd.Env, fmt.Sprintf("GETH_ANCIENT_RPC=%s", ipcPath))

	err = testCmd.Run()
	defer func() {
		abortChan <- os.Interrupt
	}()
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegration(t *testing.T) {

	_, err := session.NewSession()
	if err != nil {
		t.Skipf(`Preliminary S3 session creation failed: %v
		
		Valid S3 credentials are required for the integration test.
		
		By default NewSession will only load credentials from the shared credentials file (~/.aws/credentials).
		If the AWS_SDK_LOAD_CONFIG environment variable is set to a truthy value the Session will be created from the
		configuration values from the shared config (~/.aws/config) and shared credentials (~/.aws/credentials) files.
		Using the NewSessionWithOptions with SharedConfigState set to SharedConfigEnable will create the session as if the
		AWS_SDK_LOAD_CONFIG environment variable was set.
		> https://docs.aws.amazon.com/sdk-for-go/api/aws/session/`, err)
	}

	testCases := getTestCases()
	t.Logf("Found testcases: %v", testCases)

	for _, testCase := range testCases {
		t.Run(testCase, func(t *testing.T) {
			runTestCase(testCase, t)
		})
	}
}
