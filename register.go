package daybook

import "fmt"
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
	AddServices(pattern string, services []string) error
	RemoveServices(pattern string, services []string) error
	ListServices(pattern string) ([]string, error)
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

func (r *ConsulRegister) AddServices(pattern string, services []string) error {
	kv, _, err := r.client.Get(BASE_PREFIX+pattern, nil)
	if err != nil {
		return err
	}

	existing := strings.Split(string(kv.Value), ",")
	existing = append(existing, services...)

	uniques := make(map[string]struct{})
	for _, s := range existing {
		_, ok := uniques[s]
		if !ok {
			uniques[s] = struct{}{}
		}
	}

	final := make([]string, 0)
	for k, _ := range uniques {
		final = append(final, k)
	}

	newKv := &consulapi.KVPair{Key: BASE_PREFIX + pattern, Value: []byte(strings.Join(final, ","))}
	_, err = r.client.Put(newKv, nil)
	if err != nil {
		return err
	}

	return nil
}

func (r *ConsulRegister) RemoveServices(pattern string, services []string) error {
	kv, _, err := r.client.Get(BASE_PREFIX+pattern, nil)
	if err != nil {
		return err
	}

	existing := strings.Split(string(kv.Value), ",")

	uniques := make(map[string]struct{})
	for _, s := range existing {
		_, ok := uniques[s]
		if !ok {
			uniques[s] = struct{}{}
		}
	}

	for _, s := range services {
		_, ok := uniques[s]
		if ok {
			delete(uniques, s)
		}
	}

	final := make([]string, 0)
	for k, _ := range uniques {
		final = append(final, k)
	}

	newKv := &consulapi.KVPair{Key: BASE_PREFIX + pattern, Value: []byte(strings.Join(final, ","))}
	_, err = r.client.Put(newKv, nil)
	if err != nil {
		return err
	}

	return nil
}

func (r *ConsulRegister) ListServices(pattern string) ([]string, error) {
	kv, _, err := r.client.Get(BASE_PREFIX+pattern, nil)
	if err != nil {
		return nil, err
	}

	if kv == nil {
		return nil, fmt.Errorf("pattern '%s' does not exist", pattern)
	}

	return strings.Split(string(kv.Value), ","), nil
}
