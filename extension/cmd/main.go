package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"

	"github.com/skroutz/aws-lambda-secrets/tree/feature/extension/extension/pkg/extension"
)

// Constants for default values if none are supplied
const DEFAULT_TIMEOUT = 5000
const DEFAULT_REGION = "eu-central-1"
const SECRETS_FILE = "./secrets.yaml"
const OUTPUT_FILE = "./lambda-secrets.env"

// var client *http.Client
var (
	secretsFile    string
	region         string
	timeout        int
	outputFileName string

	secretsEnv map[string]string

	// extension name has to match the filename
	extensionName   = filepath.Base(os.Args[0])
	extensionClient = extension.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))
	// printPrefix     = fmt.Sprintf("[%s]", extensionName)
	// identifier  string
)

func getCommandParams() {
	// Setup command line args
	flag.IntVar(&timeout, "t", DEFAULT_TIMEOUT, "The amount of time to wait for any API call")
	flag.StringVar(&region, "r", DEFAULT_REGION, "The Amazon Region to use")
	flag.StringVar(&secretsFile, "f", SECRETS_FILE,
		"The YAML file containing SecretsManager ARNs and Env Var names")
	flag.StringVar(&outputFileName, "o", OUTPUT_FILE,
		"The file that will be populated with SecretsManager secrets as Env Vars")
	// Parse all of the command line args into the specified vars with the defaults
	flag.Parse()
}

// This function reads the secrets ARNs and names from the `secrets.yaml`` file provided with the application
func getSecretArns(secretsFile string) map[interface{}][]map[string]string {
	contents, err := ioutil.ReadFile(secretsFile)
	if err != nil {
		// log.Printf("[-] File '%s' could not be opened! %s", secretsFile, err.Error())
		panic(err)
	}

	secretArns := new(map[interface{}][]map[string]string)
	err = yaml.Unmarshal(contents, &secretArns)
	if err != nil {
		// log.Printf("[-] Could not unmarshal yaml config file %s! %s", secretsFile, err.Error())
		panic(err)
	}

	return *secretArns
}

// This function will return either an error or the retrieved and decrypted secret.
func GetSecret(ctx context.Context, cfg aws.Config, arn string) (*secretsmanager.GetSecretValueOutput, error) {
	client := secretsmanager.NewFromConfig(cfg)
	return client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(arn),
	})
}

// This function is runs as a go routine, responsible for fetching the secret from AWS Secrets Manager, identifying its type and writing it to `secretsEnv` map
func handleSecret(ctx context.Context, cfg aws.Config, secretTuple map[string]string, mtx *sync.Mutex, wg *sync.WaitGroup) {
	log.Printf("[+] Loading '%s' from '%s'\n", secretTuple["name"], secretTuple["valueFrom"])
	defer wg.Done()

	// https://go.dev/blog/defer-panic-and-recover
	// try to fetch each ARN
	result, err := GetSecret(ctx, cfg, secretTuple["valueFrom"])
	if err != nil {
		// log.Printf("[-] AWS Secret '%s' could not be loaded. %s", secretTuple["valueFrom"], err.Error())
		panic(err)
		// return
	}
	mtx.Lock()
	if result.SecretString != nil {
		secretsEnv[secretTuple["name"]] = *result.SecretString
		_, err = json.Marshal(*result.SecretString)
		if err != nil {
			secretsEnv[secretTuple["name"]+"_TYPE"] = "JSON"
		} else {
			secretsEnv[secretTuple["name"]+"_TYPE"] = "PLAIN"
		}
	} else {
		secretsEnv[secretTuple["name"]] = string(result.SecretBinary)
		secretsEnv[secretTuple["name"]+"_TYPE"] = "BINARY"
	}
	mtx.Unlock()
}

// This function initializes the AWS API context and config, `secretsEnv`, the mutex and the waitGroup, launches the go routines and waits for them to return
func fetchSecrets(secretsList []map[string]string) {
	// Setup a new context to allow for limited execution time for API calls with a default of 200 milliseconds
	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	// Load the config
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region), config.WithRetryer(func() aws.Retryer {
		// NopRetryer is used here in a global context to avoid retries on API calls
		return retry.AddWithMaxAttempts(aws.NopRetryer{}, 1)
	}))
	if err != nil {
		// log.Printf("[-] Loading AWS Default Config. %s", err.Error())
		panic(err)
	}

	// Initialize the Env map
	secretsEnv = make(map[string]string)

	// Mutex for outputFile fd
	mtx := new(sync.Mutex)
	wg := new(sync.WaitGroup)

	// Iterate through the ARNs
	for _, secretTuple := range secretsList {
		wg.Add(1)
		go handleSecret(ctx, cfg, secretTuple, mtx, wg)
	}

	// Wait for all go routines to finish
	wg.Wait()
}

func writeEnvFile(outputFileName string) {
	err := godotenv.Write(secretsEnv, outputFileName)
	if err != nil {
		panic(err)
		// log.Printf("[-] File '%s' could not be writen! %s", outputFileName, err.Error())
	}
}

func main() {
	getCommandParams()

	secretArns := getSecretArns(secretsFile)
	fetchSecrets(secretArns["secrets"])

	writeEnvFile(outputFileName)

	ctx, cancel := context.WithCancel(context.Background())
	resp, err := extensionClient.Register(ctx, extensionName)
	if err != nil {
		panic(err)
	}
	defer cancel()
}
