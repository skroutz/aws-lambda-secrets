package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
	"github.com/skroutz/aws-lambda-secrets/internal/smsecrets"
	"github.com/skroutz/aws-lambda-secrets/internal/utils"
)

// Constants for default values if none are supplied
const DEFAULT_TIMEOUT = 5000
const DEFAULT_REGION = "eu-central-1"
const DEFAULT_SECRETS_FILE = "secrets.yaml"
const DEFAULT_OUTPUT_FILE = "/tmp/lambda-secrets.env"
const DEFAULT_ENTRYPOINT_ENV_VAR = "ENTRYPOINT"

var (
	timeout          int
	region           string
	secretsFile      string
	secretsEnv       map[string]string
	outputFileName   string
	entrypointEnvVar string
	entrypointArray  []string

	sm *smsecrets.SecretsManager
)

func getCommandParams() {
	// Setup command line args
	flag.IntVar(&timeout, "t", utils.EnvOrInt("SECRETS_TIMEOUT", DEFAULT_TIMEOUT), "The amount of time to wait for any API call")
	flag.StringVar(&region, "r", utils.EnvOrString("SECRETS_AWS_REGION", DEFAULT_REGION), "The Amazon Region to use")
	flag.StringVar(&secretsFile, "f", utils.EnvOrString("SECRETS_FILE", DEFAULT_SECRETS_FILE),
		"The YAML file containing SecretsManager ARNs and Env Var names")
	flag.StringVar(&outputFileName, "o", utils.EnvOrString("SECRETS_OUTPUT_FILE", DEFAULT_OUTPUT_FILE),
		"The file that will be populated with SecretsManager secrets as Env Vars")
	flag.StringVar(&entrypointEnvVar, "e", utils.EnvOrString("ENTRYPOINT", DEFAULT_ENTRYPOINT_ENV_VAR),
		"The name of the Env Var storing the application entrypoint (Default: ENTRYPOINT)")

	// Parse all of the command line args into the specified vars with the defaults
	flag.Parse()

	if flag.NArg() != 0 {
		log.Printf("[*] Positional Argument treated as entrypoint: %s", flag.Args())
		entrypointArray = flag.Args()
	} else if os.Getenv(entrypointEnvVar) != "" {
		log.Printf("[*] Environment Variable '%s' is treated as entrypoint: %s", DEFAULT_ENTRYPOINT_ENV_VAR, os.Getenv(DEFAULT_ENTRYPOINT_ENV_VAR))
	} else {
		log.Println("[!] No entrypoint found")
	}
}

func ExecuteEntrypoint() (string, error) {
	err := godotenv.Load(outputFileName)
	if err != nil {
		log.Printf("[-] Error loading  EnvVars from '%s' file. %s", outputFileName, err.Error())
		return "", err
	}

	cmd := []byte{}
	if entrypointArray == nil {
		entrypoint := os.Getenv(entrypointEnvVar)
		log.Printf("[+] Passing execution to '%s'\n\n", entrypoint)
		cmd, err = exec.Command("sh", "-c", entrypoint).Output()
	} else {
		log.Printf("[+] Passing execution to '%s'\n\n", entrypointArray)
		cmd, err = exec.Command(entrypointArray[0], entrypointArray[1:]...).Output()
	}
	if err != nil {
		log.Printf("[-] Error running the entrypoint. '%s'", err)
		return "", err
	}

	fmt.Println(string(cmd))

	log.Printf("[+] Execution finished")
	return string(cmd), nil
}

func LoadLambdaSecrets() (string, error) {

	// Check if output file exists
	// If it does load it, pass execution and exit
	log.Printf("[*] Looking for Dotenv file '%s'", outputFileName)
	if stat, err := os.Stat(outputFileName); err == nil {
		if stat.Size() != 0 {
			log.Printf("Dotenv file '%s' found!", outputFileName)
		}
	} else {
		log.Printf("[!] Dotenv file '%s' NOT found!", outputFileName)
		log.Println("[*] Loading Secrets from AWS SecretsManager")
		sm = smsecrets.NewSecretsManager(region, timeout)
		secretArns := smsecrets.GetSecretArns(secretsFile)
		sm.FetchSecrets(secretArns["secrets"])
		smsecrets.WriteEnvFile(outputFileName)
	}

	// Now that the secrets are hopefully set
	output, err := ExecuteEntrypoint()

	return output, err

}

func main() {

	getCommandParams()

	LoadLambdaSecrets()

}
