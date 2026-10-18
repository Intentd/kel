// pkg/note/kube_ingress_test.go
package kube

import (
	"testing"
)

func makeIngress(lbIP, lbHost string, rules, tlsHosts []string) map[string]interface{} {
	lbEntry := map[string]interface{}{"ip": lbIP, "hostname": lbHost}
	ruleList := make([]interface{}, 0, len(rules))
	for _, h := range rules {
		ruleList = append(ruleList, map[string]interface{}{"host": h})
	}
	tlsList := []interface{}{}
	if len(tlsHosts) > 0 {
		hosts := make([]interface{}, 0, len(tlsHosts))
		for _, h := range tlsHosts {
			hosts = append(hosts, h)
		}
		tlsList = []interface{}{map[string]interface{}{"hosts": hosts}}
	}
	return map[string]interface{}{
		"spec": map[string]interface{}{
			"rules": ruleList,
			"tls":   tlsList,
		},
		"status": map[string]interface{}{
			"loadBalancer": map[string]interface{}{
				"ingress": []interface{}{lbEntry},
			},
		},
	}
}

func TestFunctionIngressReady(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want bool
	}{
		{"nil", nil, false},
		{"no status", map[string]interface{}{}, false},
		{"no lb", map[string]interface{}{"status": map[string]interface{}{}}, false},
		{"lb with ip", makeIngress("34.1.2.3", "", nil, nil), true},
		{"lb with hostname", makeIngress("", "abc.elb.amazonaws.com", nil, nil), true},
		{"lb empty both", makeIngress("", "", nil, nil), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IngressReady(tt.obj); got != tt.want {
				t.Errorf("IngressReady() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionIngressHost(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want string
	}{
		{"nil", nil, ""},
		{"no lb", map[string]interface{}{}, ""},
		{"with hostname", makeIngress("", "abc.elb.amazonaws.com", nil, nil), "abc.elb.amazonaws.com"},
		{"ip only no hostname", makeIngress("34.1.2.3", "", nil, nil), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IngressHost(tt.obj); got != tt.want {
				t.Errorf("IngressHost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionIngressIP(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want string
	}{
		{"nil", nil, ""},
		{"no lb", map[string]interface{}{}, ""},
		{"with ip", makeIngress("34.1.2.3", "", nil, nil), "34.1.2.3"},
		{"hostname only no ip", makeIngress("", "abc.elb.amazonaws.com", nil, nil), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IngressIP(tt.obj); got != tt.want {
				t.Errorf("IngressIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
