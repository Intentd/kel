package kube

import (
	"text/template"
)

// containerFuncs — Container
//
// Domain: container
//
// Read the containers of a workload (a Deployment, StatefulSet, Job, Pod, …,
// through its pod template): images, environment variables, ports and counts,
// by container index or name.
func containerFuncs() template.FuncMap {
	return template.FuncMap{
		"containerImage":      ContainerImage,
		"containerEnv":        ContainerEnv,
		"containerPort":       ContainerPort,
		"hasContainerPort":    HasContainerPort,
		"containerPortByName": ContainerPortByName,
		"containerPorts":      ContainerPorts,
		"containerCount":      ContainerCount,
	}
}

// ContainerImage returns the container image at the given index.
//
// Keywords: container, image, deployment, pod, version, runtime
//
// Example:
//
//	{{ containerImage .deployment }}    → "nginx:1.25"
//	{{ containerImage .deployment 1 }}  → "envoyproxy/envoy:v1.30"
func ContainerImage(obj interface{}, args ...int) string {
	idx := 0
	if len(args) > 0 {
		idx = args[0]
	}

	spec := Get(obj, "spec", "template", "spec")
	m, ok := spec.(map[string]interface{})
	if !ok {
		return ""
	}
	containers, ok := m["containers"].([]interface{})
	if !ok || idx < 0 || idx >= len(containers) {
		return ""
	}
	cm, ok := containers[idx].(map[string]interface{})
	if !ok {
		return ""
	}
	img, _ := cm["image"].(string)
	return img
}

// ContainerEnv returns the value of a named environment variable from the
// container at the given index.
//
// Keywords: container, environment, env, variable, deployment, pod, config
//
// Example:
//
//	{{ containerEnv .deployment "APP_ENV" }}  → "production"
func ContainerEnv(obj interface{}, key string, args ...int) string {
	idx := 0
	if len(args) > 0 {
		idx = args[0]
	}

	spec := Get(obj, "spec", "template", "spec")
	m, ok := spec.(map[string]interface{})
	if !ok {
		return ""
	}
	containers, ok := m["containers"].([]interface{})
	if !ok || idx < 0 || idx >= len(containers) {
		return ""
	}
	cm, ok := containers[idx].(map[string]interface{})
	if !ok {
		return ""
	}
	envList, ok := cm["env"].([]interface{})
	if !ok {
		return ""
	}
	for _, e := range envList {
		if em, ok := e.(map[string]interface{}); ok {
			if name, ok := em["name"].(string); ok && name == key {
				if val, ok := em["value"].(string); ok {
					return val
				}
			}
		}
	}
	return ""
}

// ContainerPort returns a containerPort number: by default the first port of
// the first container; an optional container index and port index choose
// another. 0 when there is no such port.
//
// Keywords: container, port, expose, deployment, pod, int, network
//
// Example:
//
//	{{ containerPort .deployment }}  → 8080
func ContainerPort(obj interface{}, indices ...int) int {
	containerIdx := 0
	portIdx := 0
	if len(indices) > 0 {
		containerIdx = indices[0]
	}
	if len(indices) > 1 {
		portIdx = indices[1]
	}

	containers := getContainers(obj)
	if containerIdx < 0 || containerIdx >= len(containers) {
		return 0
	}
	cm, ok := containers[containerIdx].(map[string]interface{})
	if !ok {
		return 0
	}
	ports, ok := cm["ports"].([]interface{})
	if !ok || portIdx < 0 || portIdx >= len(ports) {
		return 0
	}
	pm, ok := ports[portIdx].(map[string]interface{})
	if !ok {
		return 0
	}
	return int(toInt64(pm["containerPort"]))
}

// HasContainerPort reports whether the container exposes a specific port
// number.
//
// Keywords: container, has, port
//
// Example:
//
//	{{ hasContainerPort .deployment 0 9090 }}  → true
func HasContainerPort(obj interface{}, containerIndex int, port int) bool {
	containers := getContainers(obj)
	if containerIndex < 0 || containerIndex >= len(containers) {
		return false
	}
	cm, ok := containers[containerIndex].(map[string]interface{})
	if !ok {
		return false
	}
	ports, ok := cm["ports"].([]interface{})
	if !ok {
		return false
	}
	for _, p := range ports {
		pm, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		if int(toInt64(pm["containerPort"])) == port {
			return true
		}
	}
	return false
}

// ContainerPortByName returns the port number for a named container port.
// Returns 0 when not found.
//
// Keywords: by, container, name, port
//
// Example:
//
//	{{ containerPortByName .deployment 0 "metrics" }}  → 9090
func ContainerPortByName(obj interface{}, index int, name string) int {
	if name == "" {
		return 0
	}
	containers := getContainers(obj)
	if index < 0 || index >= len(containers) {
		return 0
	}
	cm, ok := containers[index].(map[string]interface{})
	if !ok {
		return 0
	}
	ports, ok := cm["ports"].([]interface{})
	if !ok {
		return 0
	}
	for _, p := range ports {
		pm, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		if n, _ := pm["name"].(string); n == name {
			return int(toInt64(pm["containerPort"]))
		}
	}
	return 0
}

// ContainerPorts returns the containerPort numbers the container at index
// exposes.
//
// Keywords: container, ports
//
// Example:
//
//	{{ containerPorts .deployment 0 }}  → [8080 9090]
func ContainerPorts(obj interface{}, index int) []int64 {
	containers := getContainers(obj)
	if index < 0 || index >= len(containers) {
		return nil
	}
	ports, _ := object(containers[index])["ports"].([]interface{})
	var out []int64
	for _, p := range ports {
		if n := toInt64(object(p)["containerPort"]); n > 0 {
			out = append(out, n)
		}
	}
	return out
}

// ContainerCount returns the number of containers in the pod spec.
//
// Keywords: container, count
//
// Example:
//
//	{{ containerCount .deployment }}  → 2
func ContainerCount(obj interface{}) int {
	return len(getContainers(obj))
}

func getContainers(obj interface{}) []interface{} {
	spec := Get(obj, "spec", "template", "spec")
	m, ok := spec.(map[string]interface{})
	if !ok {
		return nil
	}
	containers, _ := m["containers"].([]interface{})
	return containers
}
