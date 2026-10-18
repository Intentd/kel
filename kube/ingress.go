package kube

import (
	"text/template"
)

// ingressFuncs — Ingress
//
// Domain: ingress
//
// Read an Ingress: whether it has a load balancer address, its host and IP,
// its class, the hosts in its rules and TLS section, and every load balancer
// address.
func ingressFuncs() template.FuncMap {
	return template.FuncMap{
		"ingressReady":           IngressReady,
		"ingressHost":            IngressHost,
		"ingressIP":              IngressIP,
		"ingressClassName":       IngressClassName,
		"ingressRules":           IngressRules,
		"ingressTLSHosts":        IngressTLSHosts,
		"ingressLoadBalancerIPs": IngressLoadBalancerIPs,
	}
}

// IngressReady returns `true` when at least one load-balancer entry has been
// assigned — either an IP or hostname.
//
// Keywords: ingress, networking, loadbalancer, ready, boolean, provisioned, assigned
//
// Example:
//
//	{{ ingressReady .ingress }}  → true
func IngressReady(obj interface{}) bool {
	entry := ingressLBEntry(obj)
	if entry == nil {
		return false
	}
	ip, _ := entry["ip"].(string)
	host, _ := entry["hostname"].(string)
	return ip != "" || host != ""
}

// IngressHost returns `status.loadBalancer.ingress[0].hostname` — cloud
// providers such as AWS ALB and GCP assign a hostname rather than an IP.
//
// Keywords: ingress, networking, loadbalancer, hostname, external, cloud, aws, gcp, string
//
// Example:
//
//	{{ ingressHost .ingress }}  → "abc123.us-east-1.elb.amazonaws.com"
func IngressHost(obj interface{}) string {
	entry := ingressLBEntry(obj)
	if entry == nil {
		return ""
	}
	v, _ := entry["hostname"].(string)
	return v
}

// IngressIP returns `status.loadBalancer.ingress[0].ip`, or `""` while the LB
// is provisioning.
//
// Keywords: ingress, networking, loadbalancer, ip, address, external, string
//
// Example:
//
//	{{ ingressIP .ingress }}  → "34.123.45.67"
func IngressIP(obj interface{}) string {
	entry := ingressLBEntry(obj)
	if entry == nil {
		return ""
	}
	v, _ := entry["ip"].(string)
	return v
}

// IngressClassName returns `spec.ingressClassName`, or `""` when not set.
//
// Keywords: ingress, networking, class, string, controller, nginx, traefik
//
// Example:
//
//	{{ ingressClassName .ingress }}  → "nginx"
func IngressClassName(obj interface{}) string {
	spec := Spec(obj)
	if spec == nil {
		return ""
	}
	v, _ := spec["ingressClassName"].(string)
	return v
}

// IngressRules returns the hostnames in spec.rules, in order.
// Empty hosts (catch-all rules) are omitted.
//
// Keywords: ingress, networking, rules, hosts, routing, string, list
//
// Example:
//
//	{{ ingressRules .ingress }}  → [api.example.com www.example.com]
func IngressRules(obj interface{}) []string {
	spec := Spec(obj)
	if spec == nil {
		return nil
	}
	rules, _ := spec["rules"].([]interface{})
	var hosts []string
	for _, r := range rules {
		rm, _ := r.(map[string]interface{})
		if rm == nil {
			continue
		}
		if h, _ := rm["host"].(string); h != "" {
			hosts = append(hosts, h)
		}
	}
	return hosts
}

// IngressTLSHosts returns the distinct TLS hostnames in spec.tls.
//
// Keywords: ingress, networking, tls, hosts, ssl, string, list, https
//
// Example:
//
//	{{ ingressTLSHosts .ingress }}  → [api.example.com]
func IngressTLSHosts(obj interface{}) []string {
	spec := Spec(obj)
	if spec == nil {
		return nil
	}
	tlsEntries, _ := spec["tls"].([]interface{})
	var hosts []string
	seen := map[string]bool{}
	for _, t := range tlsEntries {
		tm, _ := t.(map[string]interface{})
		if tm == nil {
			continue
		}
		tlsHosts, _ := tm["hosts"].([]interface{})
		for _, h := range tlsHosts {
			if s, _ := h.(string); s != "" && !seen[s] {
				seen[s] = true
				hosts = append(hosts, s)
			}
		}
	}
	return hosts
}

// IngressLoadBalancerIPs returns each load balancer address in
// status.loadBalancer.ingress: its IP, or its hostname when it has no IP.
//
// Keywords: ingress, networking, loadbalancer, addresses, string, list
//
// Example:
//
//	{{ ingressLoadBalancerIPs .ingress }}  → [34.123.45.67]
func IngressLoadBalancerIPs(ingress interface{}) []string {
	entries, _ := nested(object(ingress), "status", "loadBalancer", "ingress").([]interface{})
	var out []string
	for _, e := range entries {
		em := object(e)
		if ip := nestedString(em, "ip"); ip != "" {
			out = append(out, ip)
		} else if h := nestedString(em, "hostname"); h != "" {
			out = append(out, h)
		}
	}
	return out
}

// Helpers
func ingressLBEntry(obj interface{}) map[string]interface{} {
	status := Status(obj)
	lb, _ := status["loadBalancer"].(map[string]interface{})
	if lb == nil {
		return nil
	}
	ingress, _ := lb["ingress"].([]interface{})
	if len(ingress) == 0 {
		return nil
	}
	m, _ := ingress[0].(map[string]interface{})
	return m
}
