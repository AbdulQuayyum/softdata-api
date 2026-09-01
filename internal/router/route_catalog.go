package router

import (
	"fmt"
	"net/http"
	"strings"
)

type routeCatalog struct {
	routes []routePattern
}

type routePattern struct {
	method string
	path   string
}

func (c *routeCatalog) add(pattern string) error {
	method, path, ok := strings.Cut(strings.TrimSpace(pattern), " ")
	if !ok {
		return fmt.Errorf("invalid route pattern %q", pattern)
	}
	method = strings.TrimSpace(method)
	path = strings.TrimSpace(path)
	if method == "" || path == "" {
		return fmt.Errorf("invalid route pattern %q", pattern)
	}

	c.routes = append(c.routes, routePattern{method: method, path: path})
	return nil
}

func (c *routeCatalog) allow(path string) []string {
	if c == nil || strings.TrimSpace(path) == "" {
		return nil
	}

	allow := make([]string, 0, 4)
	for _, route := range c.routes {
		if !route.matchesPath(path) {
			continue
		}
		if route.method == http.MethodHead {
			continue
		}
		if containsMethod(allow, route.method) {
			continue
		}
		allow = append(allow, route.method)
	}
	return allow
}

func (c *routeCatalog) supports(method, path string) bool {
	if c == nil {
		return false
	}
	for _, route := range c.routes {
		if route.method == method && route.matchesPath(path) {
			return true
		}
	}
	return false
}

func (r routePattern) matchesPath(path string) bool {
	templateParts := splitRoutePath(r.path)
	pathParts := splitRoutePath(path)
	for i := 0; i < len(templateParts); i++ {
		if isRouteCatchAll(templateParts[i]) {
			if i != len(templateParts)-1 {
				return false
			}
			if len(pathParts) <= i {
				return false
			}
			for _, part := range pathParts[i:] {
				if part == "" {
					return false
				}
			}
			return true
		}
		if i >= len(pathParts) {
			return false
		}
		if !routeSegmentMatches(templateParts[i], pathParts[i]) {
			return false
		}
	}
	return len(templateParts) == len(pathParts)
}

func splitRoutePath(path string) []string {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{""}
	}
	return strings.Split(path, "/")
}

func isRouteParam(segment string) bool {
	return strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") && len(segment) > 2
}

func isRouteCatchAll(segment string) bool {
	return strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "...}") && len(segment) > len("{...}")
}

func routeSegmentMatches(templatePart, pathPart string) bool {
	if isRouteCatchAll(templatePart) {
		return pathPart != ""
	}
	if isRouteParam(templatePart) {
		return pathPart != ""
	}

	if _, suffix, ok := splitRouteParamSuffix(templatePart); ok {
		if pathPart == "" || suffix == "" || !strings.HasSuffix(pathPart, suffix) {
			return false
		}
		return len(pathPart) > len(suffix)
	}

	return templatePart == pathPart
}

func splitRouteParamSuffix(segment string) (wildcard, suffix string, ok bool) {
	if !strings.HasPrefix(segment, "{") {
		return "", "", false
	}
	closeIdx := strings.IndexByte(segment, '}')
	if closeIdx <= 1 || closeIdx == len(segment)-1 {
		return "", "", false
	}
	return segment[1:closeIdx], segment[closeIdx+1:], true
}

func containsMethod(methods []string, method string) bool {
	for _, existing := range methods {
		if existing == method {
			return true
		}
	}
	return false
}
