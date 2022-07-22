package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/joho/godotenv"
)

// Constants for default values if none are supplied
const DEFAULT_TIMEOUT = 5000
const DEFAULT_REGION = "eu-central-1"
const SECRETS_FILE = "secrets.yaml"
const OUTPUT_FILE = "/tmp/lambda-secrets.env"
const ENTRYPOINT_ENV_VAR = "ENTRYPOINT"

var (
	timeout          int
	region           string
	secretsFile      string
	outputFileName   string
	entrypointEnvVar string
	entrypointArray  []string
	exitCode      	 int
)

func getCommandParams() {
	// Setup command line args
	flag.IntVar(&timeout, "t", DEFAULT_TIMEOUT, "The amount of time to wait for any API call")
	flag.StringVar(&region, "r", DEFAULT_REGION, "The Amazon Region to use")
	flag.StringVar(&secretsFile, "f", SECRETS_FILE,
		"The YAML file containing SecretsManager ARNs and Env Var names")
	flag.StringVar(&outputFileName, "o", OUTPUT_FILE,
		"The file that will be populated with SecretsManager secrets as Env Vars")
	flag.StringVar(&entrypointEnvVar, "e", ENTRYPOINT_ENV_VAR,
		"The file that will be populated with SecretsManager secrets as Env Vars")

	// Parse all of the command line args into the specified vars with the defaults
	flag.Parse()

	if flag.NArg() != 0 {
		log.Printf("[*] Positional Argument treated as entrypoint: %s", flag.Args())
		entrypointArray = flag.Args()
	} else if os.Getenv(ENTRYPOINT_ENV_VAR) != "" {
		log.Printf("[*] Environment Variable '%s' is treated as entrypoint: %s", ENTRYPOINT_ENV_VAR, os.Getenv(ENTRYPOINT_ENV_VAR))
	} else {
		log.Println("[!] No entrypoint found")
	}
}

// This function will return either an error or the retrieved and decrypted secret.
func GetSecret(ctx context.Context, cfg aws.Config, arn string) (*secretsmanager.GetSecretValueOutput, error) {
	client := secretsmanager.NewFromConfig(cfg)
	return client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(arn),
	})
}

// A function that states the passed values as dotenv format with 'export'
func CreateExportLine(envvar string, secret string) string {
	escaped_secret := strings.Replace(secret, "\"", "\\\"", -1)
	result := fmt.Sprintf("export %s=\"%s\"\n", envvar, escaped_secret)
	return result
}

func handleSecret(ctx context.Context, cfg aws.Config, secretTuple map[string]string, outputFile *os.File, mtx *sync.Mutex, wg *sync.WaitGroup) {
	log.Printf("[+] Loading '%s' from '%s'\n", secretTuple["name"], secretTuple["valueFrom"])
	defer wg.Done()

	// try to fetch each ARN
	result, err := GetSecret(ctx, cfg, secretTuple["valueFrom"])
	if err != nil {
		log.Printf("[-] AWS Secret '%s' could not be loaded. %s", secretTuple["valueFrom"], err.Error())
		exitCode = 101
		return
	}
	exportLine := CreateExportLine(secretTuple["name"], *result.SecretString)

	mtx.Lock()
	_, err = outputFile.Write([]byte(exportLine))
	mtx.Unlock()
	if err != nil {
		log.Printf("[-] Error Writing to File: %s", outputFileName)
		exitCode = 4
		return
	}
}

// This function starts execution of the entrypoint
// and exits when it returns
func ExecuteEntrypoint() {
	err := godotenv.Load(outputFileName)
	if err != nil {
		log.Printf("[-] Error loading  EnvVars from '%s' file", outputFileName)
		os.Exit(200)
	}
	err = nil
	cmd := []byte{}
	if entrypointArray == nil {
		entrypoint := os.Getenv(ENTRYPOINT_ENV_VAR)
		log.Printf("[+] Passing execution to '%s'\n\n", entrypoint)
		cmd, err = exec.Command("sh", "-c", entrypoint).Output()
	} else {
		log.Printf("[+] Passing execution to '%s'\n\n", entrypointArray)
		cmd, err = exec.Command(entrypointArray[0], entrypointArray[1:]...).Output()
	}
	if err != nil {
		log.Printf("[-] Error running the entrypoint. '%s'", err)
		os.Exit(201)
		return
	}

	fmt.Println(string(cmd))

	log.Printf("[+] Execution finished")
	os.Exit(0)
}

func main() {

	// ================
	// Get all of the command line data and perform the necessary validation
	getCommandParams()

	// Check if output file exists
	// If it does load it, pass execution and exit
	log.Printf("[*] Looking for Dotenv file '%s'", outputFileName)
	if stat, err := os.Stat(outputFileName); err == nil {
		if stat.Size() != 0 {
			log.Printf("Dotenv file '%s' found!", outputFileName)
			ExecuteEntrypoint()
		}
	}

	log.Printf("[!] Dotenv file '%s' NOT found!", outputFileName)
	log.Println("[*] Loading Secrets from AWS SecretsManager")

	// Setup a new context to allow for limited execution time for API calls with a default of 200 milliseconds
	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	// Load the config
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region), config.WithRetryer(func() aws.Retryer {
		// NopRetryer is used here in a global context to avoid retries on API calls
		return retry.AddWithMaxAttempts(aws.NopRetryer{}, 1)
	}))
	if err != nil {
		log.Println("[*] Loading Secrets from AWS SecretsManager")
	}

	// ================
	// Read the file contents
	content, err := ioutil.ReadFile(secretsFile)
	if err != nil {
		log.Printf("[-] File '%s' could not be opened!", secretsFile)
		os.Exit(1)
	}

	// Parse the file - YAML
	secretArnStruct := make(map[interface{}][]map[string]string)
	err = yaml.Unmarshal(content, secretArnStruct)
	if err != nil {
		log.Printf("[-] File '%s' could not be parsed!", secretsFile)
		os.Exit(2)
	}

	// Open the output file for writing
	outputFile, err := os.OpenFile(outputFileName,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		log.Printf("[-] File '%s' could not be writen!", outputFileName)
		os.Exit(3)
	}

	// Mutex for outputFile fd
	mtx := new(sync.Mutex)
	wg := new(sync.WaitGroup)

	// Iterate through the ARNs
	secretsList := secretArnStruct["secrets"]
	for _, secretTuple := range secretsList {
		wg.Add(1)
		go handleSecret(ctx, cfg, secretTuple, outputFile, mtx, wg)
	}


	// Wait for all go routines to finish
	wg.Wait()
	outputFile.Close()
	if exitCode != 0{
		os.Exit(exitCode)
	}

	// Now that the secrets are set
	// Pass execution
	ExecuteEntrypoint()
}
