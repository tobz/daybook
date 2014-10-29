package main

import "os"
import "log"
import "github.com/mitchellh/goamz/aws"
import "github.com/tobz/daybook"

func main() {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Failed to get hostname: %s", err)
	}

	register, err := daybook.NewConsulRegister()
	if err != nil {
		log.Fatalf("Failed to get Consul register: %s", err)
	}

	auth, err := aws.GetAuth("", "")
	if err != nil {
		log.Fatalf("Failed to get AWS credentials from environment: %s", err)
	}

	store := daybook.NewS3Store(auth, aws.USEast, "daybook")

	log.Printf("Getting services from register...")

	services, err := register.GetServices(hostname)
	if err != nil {
		log.Fatalf("Got an error while retreiving a list of service mappings: %s", err)
	}

	log.Printf("Got %d service(s) from register!", len(services))

	for _, service := range services {
		log.Printf("Getting assets for service '%s'...", service.Name)

		assets, err := store.GetAll(service.Name)
		if err != nil {
			log.Fatalf("Got an error while grabbing assets for service '%s': %s", service.Name, err)
		}

		for _, asset := range assets {
			log.Printf("Pulling down %s/%s...", asset.Name, asset.Version)

			archive, err := store.Get(asset)
			if err != nil {
				log.Fatalf("Got an error while grabbing asset %s/%s: %s", asset.Name, asset.Version, err)
			}

			if err = archive.Extract("."); err != nil {
				log.Fatalf("Got an error while extracting service asset %s/%s: %#v", asset.Name, asset.Version, err)
			}
		}
	}

	log.Print("All done!")
}
