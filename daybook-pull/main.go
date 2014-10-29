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
	Hostname         string `json:"hostname"`
	InstallDirectory string `json:"install_dir"`
	BucketName       string `json:"bucket_name"`
	AWSAccessKey     string `json:"aws_access_key"`
	AWSSecretKey     string `json:"aws_secret_key"`
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

	if conf.Hostname == "" {
		conf.Hostname, err = os.Hostname()
		if err != nil {
			log.Fatalf("Failed to get hostname from configuration and environment: %s", err)
		}
	}

	if conf.BucketName == "" {
		conf.BucketName = "daybook"
	}

	if conf.InstallDirectory == "" {
		conf.InstallDirectory = "/tmp"
	}

	register, err := daybook.NewConsulRegister()
	if err != nil {
		log.Fatalf("Failed to get Consul register: %s", err)
	}

	auth, err := aws.GetAuth(conf.AWSAccessKey, conf.AWSSecretKey)
	if err != nil {
		log.Fatalf("Failed to get AWS credentials from environment: %s", err)
	}

	store := daybook.NewS3Store(auth, aws.USEast, conf.BucketName)

	log.Printf("Getting services from register...")

	services, err := register.GetServices(conf.Hostname)
	if err != nil {
		log.Fatalf("Failed to retreive a list of service mappings: %s", err)
	}

	log.Printf("Got %d service(s) from register!", len(services))

	for _, service := range services {
		log.Printf("Getting assets for service '%s'...", service.Name)

		assets, err := store.GetAll(service.Name)
		if err != nil {
			log.Fatalf("Failed to grab list of assets for service '%s': %s", service.Name, err)
		}

		for _, asset := range assets {
			log.Printf("Pulling down %s/%s...", asset.Name, asset.Version)

			archive, err := store.Get(asset)
			if err != nil {
				log.Fatalf("Failed to grab asset %s/%s: %s", asset.Name, asset.Version, err)
			}

			installPath := filepath.Join(conf.InstallDirectory, asset.Name, asset.Version)

			err = os.MkdirAll(installPath, 0755)
			if err != nil {
				log.Fatalf("Failed to create parent directory '%s' prior to asset extraction: %s", installPath, err)
			}

			if err = archive.Extract(installPath); err != nil {
				log.Fatalf("Failed to extract service asset %s/%s: %#v", asset.Name, asset.Version, err)
			}
		}
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
