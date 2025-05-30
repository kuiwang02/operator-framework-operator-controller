package olmv1util

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// PolicyRule represents a Kubernetes RBAC policy rule
type PolicyRule struct {
	APIGroups       []string `yaml:"apiGroups,omitempty"`
	Resources       []string `yaml:"resources,omitempty"`
	ResourceNames   []string `yaml:"resourceNames,omitempty"`
	NonResourceURLs []string `yaml:"nonResourceURLs,omitempty"`
	Verbs           []string `yaml:"verbs"`
}

// Metadata holds metadata for Role/ClusterRole
type Metadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace,omitempty"`
}

// ClusterRole manifest definition
type ClusterRole struct {
	APIVersion string       `yaml:"apiVersion"`
	Kind       string       `yaml:"kind"`
	Metadata   Metadata     `yaml:"metadata"`
	Rules      []PolicyRule `yaml:"rules"`
}

// Role manifest definition
type Role struct {
	APIVersion string       `yaml:"apiVersion"`
	Kind       string       `yaml:"kind"`
	Metadata   Metadata     `yaml:"metadata"`
	Rules      []PolicyRule `yaml:"rules"`
}

func GenerateRBACFromMissingRules(missingrules, cename, roleDir string) error {
	lines := []string{}
	for _, ln := range strings.Split(missingrules, "\n") {
		if t := strings.TrimSpace(ln); t != "" {
			lines = append(lines, t)
		}
	}
	if len(lines) == 0 {
		return fmt.Errorf("no missing rules provided")
	}

	clusterRules := []PolicyRule{}
	namespaced := map[string][]PolicyRule{}
	for _, line := range lines {
		ns, rule := parseRule(line)
		if rule == nil {
			continue
		}
		if ns == "" {
			clusterRules = append(clusterRules, *rule)
		} else {
			namespaced[ns] = append(namespaced[ns], *rule)
		}
	}

	if len(clusterRules) == 0 {
		return fmt.Errorf("no cluster-scoped rules parsed")
	}

	cr := ClusterRole{
		APIVersion: "rbac.authorization.k8s.io/v1",
		Kind:       "ClusterRole",
		Metadata:   Metadata{Name: cename + "-clusterrole"},
		Rules:      clusterRules,
	}
	out := fmt.Sprintf("%s/%s-clusterrole.yaml", roleDir, cename)
	if err := writeYAML(out, cr); err != nil {
		return fmt.Errorf("write clusterrole: %w", err)
	}

	for ns, rules := range namespaced {
		if len(rules) == 0 {
			continue
		}
		role := Role{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "Role",
			Metadata:   Metadata{Name: fmt.Sprintf("%s-role-%s", cename, ns), Namespace: ns},
			Rules:      rules,
		}
		out := fmt.Sprintf("%s/%s-role-%s.yaml", roleDir, cename, ns)
		if err := writeYAML(out, role); err != nil {
			return fmt.Errorf("write role for ns %s: %w", ns, err)
		}
	}
	return nil
}

func parseRule(line string) (string, *PolicyRule) {
	ns := ""
	if i := strings.Index(line, `Namespace:"`); i >= 0 {
		rest := line[i+len(`Namespace:"`):]
		if j := strings.Index(rest, `"`); j >= 0 {
			ns = rest[:j]
		}
	}
	if idx := strings.Index(line, "NonResourceURLs:"); idx >= 0 {
		urls := extractList(line[idx:], "[", "]")
		verbs := extractList(line, "Verbs:[", "]")
		return ns, &PolicyRule{NonResourceURLs: urls, Verbs: verbs}
	}
	ag := extractList(line, "APIGroups:[", "]")
	if len(ag) == 0 {
		ag = []string{""}
	}
	res := extractList(line, "Resources:[", "]")
	rn := extractList(line, "ResourceNames:[", "]")
	vs := extractList(line, "Verbs:[", "]")
	return ns, &PolicyRule{APIGroups: ag, Resources: res, ResourceNames: rn, Verbs: vs}
}

func extractList(s, start, end string) []string {
	if i := strings.Index(s, start); i >= 0 {
		sub := s[i+len(start):]
		j := strings.Index(sub, end)
		if j < 0 {
			j = len(sub)
		}
		parts := strings.Split(sub[:j], ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if t := strings.TrimSpace(p); t != "" {
				out = append(out, t)
			}
		}
		return out
	}
	return nil
}

func writeYAML(filename string, obj interface{}) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	return enc.Encode(obj)
}
