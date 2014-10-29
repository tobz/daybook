package daybook

import "log"
import "sort"
import "path"
import "strings"
import "github.com/armon/consul-api"

var BASE_PREFIX string = "daybook/hosts/"

type SpecifierList []*consulapi.KVPair

func (sl SpecifierList) Len() int {
	return len(sl)
}

func (sl SpecifierList) Less(i, j int) bool {
	return len(sl[i].Key) > len(sl[j].Key)
}

func (sl SpecifierList) Swap(i, j int) {
	sl[i], sl[j] = sl[j], sl[i]
}

type Register interface {
	GetServices(host string) ([]*Service, error)
}

type ConsulRegister struct {
	client *consulapi.KV
}

func NewConsulRegister() (*ConsulRegister, error) {
	client, err := consulapi.NewClient(consulapi.DefaultConfig())
	if err != nil {
		return nil, err
	}

	return &ConsulRegister{client: client.KV()}, nil
}

func (r *ConsulRegister) GetServices(host string) ([]*Service, error) {
	keys, _, err := r.client.List(BASE_PREFIX, nil)
	if err != nil {
		return nil, err
	}

	matches := SpecifierList{}

	for _, kv := range keys {
		pattern := kv.Key[len(BASE_PREFIX):]

		matched, err := path.Match(pattern, host)
		if err != nil {
			log.Printf("Failed to use '%s' as a matcher: %s.  Continuing...", pattern, err)
			continue
		}

		if matched {
			matches = append(matches, kv)
		}
	}

	if len(matches) > 0 {
		sort.Sort(matches)

		services := []*Service{}
		split := strings.Split(string(matches[0].Value), ",")
		for _, s := range split {
			services = append(services, &Service{Name: strings.TrimSpace(s)})
		}

		return services, nil
	}

	return nil, nil
}
