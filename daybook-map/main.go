package main

import "log"
import "flag"
import "strings"
import "github.com/tobz/daybook"

func main() {
	flag.Parse()

	register, err := daybook.NewConsulRegister()
	if err != nil {
		log.Fatalf("Failed to get Consul register: %s", err)
	}

    args := flag.Args()

    if args[0] != "add" && args[0] != "remove" && args[0] != "list" {
        log.Fatalf("Unknown action '%s'.  Valid actions: 'add', 'remove', and 'list'.", args[0])
    }

    if (args[0] == "add" || args[0] == "remove") && len(args) < 3 {
        log.Fatal("You must specify a pattern and list of services to add/remove! e.g. daybook-map add \"web-prod-*\" new_service [new_service_2 [...]]")
    }

    if args[0] == "list" && len(args) != 2 {
        log.Fatal("You must specify a pattern to list! e.g. daybook-map list \"web-prod-\\*\"")
    }

    switch args[0] {
    case "list":
        log.Printf("Getting services for '%s'...", args[1])

        services, err := register.ListServices(args[1])
        if err != nil {
            log.Fatalf("Failed to list the services for '%s': %s", args[1], err)
        }

        log.Printf("Services for '%s': %s", args[1], strings.Join(services, ", "))
    case "add":
        log.Printf("Adding services to '%s'...", args[1])

        err = register.AddServices(args[1], args[2:])
        if err != nil {
            log.Fatalf("Failed to add new services to '%s': %s", args[1], err)
        }

        services, err := register.ListServices(args[1])
        if err != nil {
            log.Fatalf("Failed to list the services for '%s': %s", args[1], err)
        }

        log.Printf("Services for '%s': %s", args[1], strings.Join(services, ", "))
    case "remove":
        log.Printf("Removing services from '%s'...", args[1])

        err = register.RemoveServices(args[1], args[2:])
        if err != nil {
            log.Fatalf("Failed to remove services from '%s': %s", args[1], err)
        }

        services, err := register.ListServices(args[1])
        if err != nil {
            log.Fatalf("Failed to list the services for '%s': %s", args[1], err)
        }

        log.Printf("Services for '%s': %s", args[1], strings.Join(services, ", "))
    }

	log.Print("All done!")
}
