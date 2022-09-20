package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	smsecrets "github.com/skroutz/aws-lambda-secrets/internal/smsecrets"
	extension "github.com/skroutz/aws-lambda-secrets/pkg/extension"
)

// Constants for default values if none are supplied
const DEFAULT_TIMEOUT = 5000
const DEFAULT_REGION = "eu-central-1"
const SECRETS_FILE = "/var/task/secrets.yaml"
const OUTPUT_FILE = "/tmp/lambda-secrets.env"
const SECRETS_CACHE_ENABLED = true
const SECRETS_CACHE_TTL = 10

var (
	secretsFile    string
	region         string
	timeout        int
	outputFileName string
	cacheTTL       time.Duration

	// Cache init
	cacheEnabled = true

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
	flag.DurationVar(&cacheTTL, "c", SECRETS_CACHE_TTL,
		"The cache TTL that defines the time window to re-fetch secrets. Setting it to < 0 will disable caching.")
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
			log.Println("[extension] Waiting for event...")
			res, err := extensionClient.NextEvent(ctx)
			if err != nil {
				panic(err)
			}
			log.Println("[extension] Received event:", res.EventType)
			// Exit on SHUTDOWN event
			if res.EventType == extension.Shutdown {
				log.Println("[extension] Received SHUTDOWN event")
				log.Println("[extension] Exiting ...")
				return
			}
		}
	}
}

func writeSecretsOnDisk(sm *smsecrets.SecretsManager) {
	secretArns := smsecrets.GetSecretArns(secretsFile)
	sm.FetchSecrets(secretArns["secrets"])
	smsecrets.WriteEnvFile(outputFileName)
}

// This function is only invoked on cold starts
func main() {

	log.Println("[extension] This function is only invoked on cold starts")
	getCommandParams()

	sm = smsecrets.NewSecretsManager(region, timeout)
	writeSecretsOnDisk(sm)

	// Lambda API client context
	ctx, cancel := context.WithCancel(context.Background())

	// Enable cache
	if cacheTTL <= 0 {
		cacheEnabled = false
	}

	// Re-Fetch cached secrets on cache TTL timeout
	var ticker *time.Ticker
	if cacheEnabled {
		ticker = time.NewTicker(cacheTTL * time.Minute)
		go func() {
			c := <-ticker.C
			log.Printf("[extension] Cache Timeout: %s. Refreshing secrets !\n", c)
			writeSecretsOnDisk(sm)
		}()
	}

	// Handle OS signals to cancel the context before terminating the extension process
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-interruptChannel
		if cacheEnabled {
			ticker.Stop()
		}
		log.Println("[extension] Received Signal: ", s)
		log.Println("[extension] Exiting")
		cancel()
	}()

	// Register extension to Lambda Runtime API
	_, err := extensionClient.Register(ctx, extensionName)
	if err != nil {
		panic(err)
	}
	// log.Println("[extension] Register Response: ", reg_resp)

	processEvents(ctx)
}
