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
	if len(templateParts) != len(pathParts) {
		return false
	}

	for i := range templateParts {
		templatePart := templateParts[i]
		if isRouteParam(templatePart) {
			if pathParts[i] == "" {
				return false
			}
			continue
		}
		if templatePart != pathParts[i] {
			return false
		}
	}
	return true
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

func containsMethod(methods []string, method string) bool {
	for _, existing := range methods {
		if existing == method {
			return true
		}
	}
	return false
}
