package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	smsecrets "github.com/skroutz/aws-lambda-secrets/internal/smsecrets"
	extension "github.com/skroutz/aws-lambda-secrets/pkg/extension"
)

// Constants for default values if none are supplied
const DEFAULT_TIMEOUT = 5000
const DEFAULT_REGION = "eu-central-1"
const SECRETS_FILE = "/var/task/secrets.yaml"
const OUTPUT_FILE = "/tmp/lambda-secrets.env"

var (
	secretsFile    string
	region         string
	timeout        int
	outputFileName string

	// extension name has to match the filename
	extensionName   = filepath.Base(os.Args[0])
	extensionClient = extension.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))
	sm              *smsecrets.SecretsManager
)

// This function parses extension parameters as CLI arguments
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

func processEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// Quit events loop on context cancellation
			return
		default:
			log.Println("Waiting for event...")
			res, err := extensionClient.NextEvent(ctx)
			if err != nil {
				panic(err)
			}
			log.Println("Received event:", res)
			// Exit on SHUTDOWN event
			if res.EventType == extension.Shutdown {
				log.Println("Received SHUTDOWN event")
				log.Println("Exiting ...")
				return
			}
		}
	}
}

// This function is only invoked on cold starts
func main() {

	log.Println("[extension] This function is only invoked on cold starts")
	getCommandParams()

	sm = smsecrets.NewSecretsManager(region, timeout)
	secretArns := smsecrets.GetSecretArns(secretsFile)
	sm.FetchSecrets(secretArns["secrets"])
	smsecrets.WriteEnvFile(outputFileName)

	// Lambda API client context
	ctx, cancel := context.WithCancel(context.Background())
	// Handle OS signals to cancel the context before terminating the extension process
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-interruptChannel
		log.Println("Received Signal: ", s)
		log.Println("Exiting")
		cancel()
	}()

	// Register extension to Lambda Runtime API
	reg_resp, err := extensionClient.Register(ctx, extensionName)
	if err != nil {
		panic(err)
	}
	log.Println("[extension] Register Response: ", reg_resp)

	processEvents(ctx)
}
