package kube

import (
	"fmt"
	"text/template"
)

// serviceFuncs — Service
//
// Domain: service
//
// Read a Service (cluster IP, node port, load balancer IP and host) and its
// EndpointSlices, passed as one slice or a list: whether any endpoint is
// ready, and the endpoints as "ip:port".
func serviceFuncs() template.FuncMap {
	return template.FuncMap{
		"serviceClusterIP":        ServiceClusterIP,
		"serviceNodePort":         ServiceNodePort,
		"serviceLoadBalancerIP":   ServiceLoadBalancerIP,
		"serviceLoadBalancerHost": ServiceLoadBalancerHost,
		"endpointsReady":          EndpointsReady,
		"hasEndpoints":            HasEndpoints,
		"serviceEndpoints":        ServiceEndpoints,
		"serviceEndpointCount":    ServiceEndpointCount,
		"serviceFirstEndpoint":    ServiceFirstEndpoint,
	}
}

// ── Endpoints of a Service, from its EndpointSlices ───────────────────────

// HasEndpoints reports whether any endpoint in the EndpointSlices is ready.
//
// Keywords: service, endpoints, ready, boolean, traffic, networking
//
// Example:
//
//	{{ hasEndpoints .endpointSlices }}
func HasEndpoints(slices interface{}) bool {
	for _, e := range endpointEntries(slices) {
		if e.ready {
			return true
		}
	}
	return false
}

// ServiceEndpoints returns the ready endpoints as "ip:port", one per address
// and port.
//
// Keywords: service, endpoints, list, ip, port, networking, addresses
//
// Example:
//
//	{{ serviceEndpoints .endpointSlices }}  → [10.0.0.1:8080 10.0.0.2:8080]
func ServiceEndpoints(slices interface{}) []string {
	var out []string
	for _, e := range endpointEntries(slices) {
		if e.ready {
			out = append(out, e.String())
		}
	}
	return out
}

// ServiceEndpointCount returns the number of endpoints (address and port
// pairs), ready or not.
//
// Keywords: service, endpoints, count, int, networking
//
// Example:
//
//	{{ serviceEndpointCount .endpointSlices }}  → 3
func ServiceEndpointCount(slices interface{}) int {
	return len(endpointEntries(slices))
}

// ServiceFirstEndpoint returns the first ready endpoint as "ip:port"; "" when
// none is ready.
//
// Keywords: service, endpoints, first, ip, port, networking, primary
//
// Example:
//
//	{{ serviceFirstEndpoint .endpointSlices }}  → "10.0.0.1:8080"
func ServiceFirstEndpoint(slices interface{}) string {
	for _, e := range endpointEntries(slices) {
		if e.ready {
			return e.String()
		}
	}
	return ""
}

type endpoint struct {
	ip    string
	port  int64
	ready bool
}

func (e endpoint) String() string { return fmt.Sprintf("%s:%d", e.ip, e.port) }

// endpointEntries expands EndpointSlices (one, a list, or an
// EndpointSliceList) into one entry per address and port.
func endpointEntries(v interface{}) []endpoint {
	slices := objectList(v)
	if m := object(v); m != nil && m["endpoints"] != nil {
		slices = []map[string]interface{}{m}
	}
	var out []endpoint
	for _, slice := range slices {
		var ports []int64
		ps, _ := slice["ports"].([]interface{})
		for _, p := range ps {
			if port := toInt64(object(p)["port"]); port > 0 {
				ports = append(ports, port)
			}
		}
		eps, _ := slice["endpoints"].([]interface{})
		for _, e := range eps {
			em := object(e)
			ready, _ := nested(em, "conditions", "ready").(bool)
			addrs, _ := em["addresses"].([]interface{})
			for _, a := range addrs {
				ip, _ := a.(string)
				if ip == "" {
					continue
				}
				for _, port := range ports {
					out = append(out, endpoint{ip, port, ready})
				}
			}
		}
	}
	return out
}

// ServiceClusterIP returns `spec.clusterIP`. Returns `""` before the IP is
// assigned or when the Service is absent.
//
// Keywords: service, networking, cluster, ip, address, internal
//
// Example:
//
//	{{ serviceClusterIP .service }}  → "10.96.14.2"
func ServiceClusterIP(obj interface{}) string {
	spec := Get(obj, "spec")
	m, ok := spec.(map[string]interface{})
	if !ok {
		return ""
	}
	v, _ := m["clusterIP"].(string)
	return v
}

// ServiceNodePort returns the `nodePort` of the first Service port. Returns
// `0` before the port is assigned.
//
// Keywords: service, networking, nodeport, port, int, external
//
// Example:
//
//	{{ serviceNodePort .service }}  → 31234
func ServiceNodePort(obj interface{}) int {
	spec := Get(obj, "spec")
	m, ok := spec.(map[string]interface{})
	if !ok {
		return 0
	}
	ports, ok := m["ports"].([]interface{})
	if !ok || len(ports) == 0 {
		return 0
	}
	pm, ok := ports[0].(map[string]interface{})
	if !ok {
		return 0
	}
	return int(toInt64(pm["nodePort"]))
}

// ServiceLoadBalancerIP returns the external IP assigned by the load balancer
// (`status.loadBalancer.ingress[0].ip`). Returns `""` while provisioning —
// cloud LB provisioning typically takes 30–90 seconds.
//
// Keywords: service, networking, loadbalancer, ip, external, cloud, provisioning
//
// Example:
//
//	{{ serviceLoadBalancerIP .service }}  → "34.123.45.67"
func ServiceLoadBalancerIP(obj interface{}) string {
	ingress := lbIngress(obj)
	if ingress == nil {
		return ""
	}
	ip, _ := ingress["ip"].(string)
	return ip
}

// ServiceLoadBalancerHost returns the hostname assigned by the load balancer
// (`status.loadBalancer.ingress[0].hostname`). Cloud providers (AWS, GCP)
// typically assign a hostname rather than an IP.
//
// Keywords: service, networking, loadbalancer, hostname, external, cloud, aws, gcp
//
// Example:
//
//	{{ serviceLoadBalancerHost .service }}  → "abc123.us-east-1.elb.amazonaws.com"
//	https://{{ serviceLoadBalancerHost .service }}
func ServiceLoadBalancerHost(obj interface{}) string {
	ingress := lbIngress(obj)
	if ingress == nil {
		return ""
	}
	host, _ := ingress["hostname"].(string)
	return host
}

func lbIngress(obj interface{}) map[string]interface{} {
	status := Status(obj)
	lb, ok := status["loadBalancer"].(map[string]interface{})
	if !ok {
		return nil
	}
	ingress, ok := lb["ingress"].([]interface{})
	if !ok || len(ingress) == 0 {
		return nil
	}
	m, _ := ingress[0].(map[string]interface{})
	return m
}

// EndpointsReady reports whether an EndpointSlice (or a core/v1 Endpoints
// object) has at least one ready address. Pass the Service's EndpointSlice,
// not the Service.
//
// Keywords: service, endpoints, ready, boolean, networking, traffic
//
// Example:
//
//	{{ endpointsReady (index .endpointSlices 0) }}  → true
func EndpointsReady(obj interface{}) bool {
	m, ok := obj.(map[string]interface{})
	if !ok {
		return false
	}

	// EndpointSlice format (discovery.k8s.io/v1)
	if eps, ok := m["endpoints"].([]interface{}); ok {
		for _, e := range eps {
			em, ok := e.(map[string]interface{})
			if !ok {
				continue
			}
			addrs, _ := em["addresses"].([]interface{})
			if len(addrs) == 0 {
				continue
			}
			cond, _ := em["conditions"].(map[string]interface{})
			ready, _ := cond["ready"].(bool)
			if ready {
				return true
			}
		}
		return false
	}

	// Legacy Endpoints format (v1): subsets at top level or under status
	for _, s := range legacyEndpointSubsets(m) {
		sm, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		addrs, ok := sm["addresses"].([]interface{})
		if ok && len(addrs) > 0 {
			return true
		}
	}
	return false
}

func legacyEndpointSubsets(m map[string]interface{}) []interface{} {
	if subsets, ok := m["subsets"].([]interface{}); ok {
		return subsets
	}
	if status, ok := m["status"].(map[string]interface{}); ok {
		if subsets, ok := status["subsets"].([]interface{}); ok {
			return subsets
		}
	}
	return nil
}
