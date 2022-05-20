package gowww

import (
	"net"
	"net/http"
	"os"
	"strings"

	"git.gohegan.uk/kaigoh/gowww/v2/utilities"
)

// Extract the "clean name" from the vhosts path...
func CleanName(path string) string {
	return strings.TrimPrefix(path, utilities.GetEnv("GOWWW_ROOT", "vhosts")+string(os.PathSeparator))
}

// Does the requested host exist?
func HaveHost(hosts map[string]string, host string) bool {
	return hosts[host] != ""
}

// Remove the requested host
func RemoveHost(hosts map[string]string, host string) (hostsUpdated map[string]string) {
	hostsUpdated = make(map[string]string)
	for k, v := range hosts {
		if v != host {
			hostsUpdated[k] = v
		}
	}
	return
}

// Does the requested host route exist?
func HaveRoute(routes []string, host string) bool {
	for _, v := range routes {
		if host == v {
			return true
		}
	}
	return false
}

func GetClientIP(r *http.Request) string {
	ip := r.Header.Get("X-REAL-IP")
	netIP := net.ParseIP(ip)
	if netIP != nil {
		return ip
	}
	ips := r.Header.Get("X-FORWARDED-FOR")
	splitIps := strings.Split(ips, ",")
	for _, ip := range splitIps {
		netIP := net.ParseIP(ip)
		if netIP != nil {
			return ip
		}
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return ""
	}
	netIP = net.ParseIP(ip)
	if netIP != nil {
		return ip
	}
	return "?.?.?.?"
}
