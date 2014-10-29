package main

import "os"
import "fmt"
import "log"
import "flag"
import "path/filepath"
import "encoding/json"
import "github.com/mitchellh/goamz/aws"
import "github.com/tobz/daybook"

var configFile = flag.String("config-file", "/etc/daybook.json", "the base Daybook configuration - AWS credentials, install root, etc")

type Configuration struct {
	BucketName   string `json:"bucket_name"`
	AWSAccessKey string `json:"aws_access_key"`
	AWSSecretKey string `json:"aws_secret_key"`
}

func main() {
	flag.Parse()

	if *configFile == "" {
		log.Fatal("You must specify a path to the Daybook configuration!")
	}

	conf, err := getConfiguration(*configFile)
	if err != nil {
		log.Fatalf("Failed to open configuration file: %s", err)
	}

	if conf.BucketName == "" {
		conf.BucketName = "daybook"
	}

	auth, err := aws.GetAuth(conf.AWSAccessKey, conf.AWSSecretKey)
	if err != nil {
		log.Fatalf("Failed to get AWS credentials from environment: %s", err)
	}

	store := daybook.NewS3Store(auth, aws.USEast, conf.BucketName)

	args := flag.Args()
	if len(args) != 3 {
		log.Fatal("Too many arguments present, try: daybook-push [config-file=/path/to/config.json] <service name> <version> <path to asset>")
	}

	log.Printf("Sending %s as %s/%s...", args[2], args[0], args[1])

	service := &daybook.Service{Name: args[0], Version: args[1]}
	err = store.Put(service, args[2])
	if err != nil {
		log.Fatalf("Failed to send service asset to S3: %s", err)
	}

	log.Print("All done!")
}

func getConfiguration(configFile string) (*Configuration, error) {
	absConfPath, err := filepath.Abs(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path to configuration: %s", err)
	}

	f, err := os.Open(absConfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open the configuration file: %s", err)
	}
	defer f.Close()

	config := &Configuration{}

	decoder := json.NewDecoder(f)
	err = decoder.Decode(config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode configuration: %s", err)
	}

	return config, nil
}
