package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/joho/godotenv"
)

// Constants for default values if none are supplied
const SECRETS_FILE = "/tmp/lambda-secrets.env"
const ENTRYPOINT_ENV_VAR = "ENTRYPOINT"

var (
	secretsFile      string
	entrypointEnvVar string
	entrypointArray  []string
)

func getCommandParams() {
	flag.StringVar(&secretsFile, "f", SECRETS_FILE, "The file populated with SecretsManager secrets as Env Vars")
	flag.StringVar(&entrypointEnvVar, "e", ENTRYPOINT_ENV_VAR, "The EnvVar that stores the lambda entrypoint")

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

func InLambda() bool {

	// These environment variables are set by AWS Lambdas,
	// used by 'aws-lambda-go' module:
	// https://github.com/aws/aws-lambda-go/blob/main/lambda/entry.go#L72
	if os.Getenv("_LAMBDA_SERVER_PORT") == "" &&
		os.Getenv("AWS_LAMBDA_RUNTIME_API") == "" {
		return false
	}
	return true
}

func LoadLambdaSecrets() {
	// Check if output file exists
	// If it does load it, pass execution and exit
	log.Printf("[*] Looking for Dotenv file '%s'", secretsFile)

	if stat, err := os.Stat(secretsFile); err == nil {
		if stat.Size() != 0 {
			log.Printf("[+] Dotenv file '%s' found!", secretsFile)
			_, err := ExecuteEntrypoint()
			if err != nil {
				panic(err)
			}
		}
	} else {
		log.Printf("[!] Dotenv file NOT '%s' found!", secretsFile)
		panic(err)
	}
}

// This function starts execution of the entrypoint and exits when it returns
func ExecuteEntrypoint() (string, error) {

	err := godotenv.Load(secretsFile)
	if err != nil {
		log.Printf("[-] Error loading  EnvVars from '%s' file. %s", secretsFile, err.Error())
		return "", err
	}

	var cmd string
	if entrypointArray == nil {
		entrypoint := os.Getenv(ENTRYPOINT_ENV_VAR)
		log.Printf("[!] entrypointArray is nil. Passing execution to '%s'\n\n", entrypoint)
		cmd, err := exec.LookPath(entrypoint)
		if err != nil {
			panic(err)
		}
		syscall.Exec(cmd, nil, os.Environ())
	} else {
		log.Printf("[!] Passing execution to '%s'\n\n", entrypointArray)
		cmd, err := exec.LookPath(entrypointArray[0])
		if err != nil {
			panic(err)
		}
		err = syscall.Exec(cmd, entrypointArray, os.Environ())
	}

	if err != nil {
		log.Printf("[-] Error running the entrypoint. '%s'", err)
		return "", err
	}

	log.Printf("[*] Execution finished")

	return cmd, nil
}

func main() {

	getCommandParams()

	if InLambda() {
		log.Println("[*] AWS Lambda Environment Detected")
		lambda.Start(LoadLambdaSecrets)
	} else {
		log.Println("[*] Not running in AWS Lambda")
		LoadLambdaSecrets()
	}
}
