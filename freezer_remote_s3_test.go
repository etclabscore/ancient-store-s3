package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	// s3BucketNamespace is used to unique-ify S3 bucket names.
	// s3 bucket names must be unique GLOBALLY.
	// This can be configured by the tester with the environment variable: AWS_BUCKET_NAMESPACE.
	// If this environment variable is not set, the OS hostname will be used.
	// Ye be warned.
	// Buckets are named with the following pattern:
	// ancientstore-<namespace>-<testname>
	//
	// Only the <namespace> value is user-configurable.
	s3BucketNamespace = ""
)

func init() {
	// AWS credential defaults.
	if os.Getenv("AWS_REGION") == "" {
		log.Println("Setting default AWS_REGION credential")
		os.Setenv("AWS_REGION", "us-west-1")
	}
	if os.Getenv("AWS_PROFILE") == "" {
		log.Println("Setting default AWS_PROFILE credential")
		os.Setenv("AWS_PROFILE", "developers-s3")
	}

	// S3 bucket naming defaults.
	s3BucketNamespace = os.Getenv("AWS_BUCKET_NAMESPACE")
	if s3BucketNamespace == "" {
		hostname, err := os.Hostname()
		if err != nil {
			panic(err)
		}
		s3BucketNamespace = hostname
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
		bucketName, err := bucketNameForTest(testCase)
		if err != nil {
			log.Fatalln("failed to create bucket name:", err)
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

	testCases := getTestCases()
	t.Logf("Found testcases: %v", testCases)

	for _, testCase := range testCases {
		t.Run(testCase, func(t *testing.T) {
			runTestCase(testCase, t)
		})
	}
}

func bucketNameForTest(testName string) (string, error) {
	// https://docs.aws.amazon.com/AmazonS3/latest/dev/BucketRestrictions.html
	//
	// These sanitizations ARE NOT COMPLETE; it is still possible for users to fuck this up.
	// But this is about the extent of my patience.
	bucketName := fmt.Sprintf("ancientstore-%s-%s", s3BucketNamespace, testName)
	bucketName = strings.ToLower(bucketName)
	re := regexp.MustCompile(`[^a-zA-Z.-]`)
	bucketName = re.ReplaceAllString(bucketName, "")
	if len(bucketName) > 63 {
		bucketName = bucketName[:63]
	}
	return bucketName, nil
}
