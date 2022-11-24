package middleware

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

// TrustedSubnetHandle allows requests for endpoint /api/internal/stats only for trusted subnet.
// This handler is used as a middleware for all server requests.
func TrustedSubnetHandle(trustedSubnet string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("r.RequestURI: %v", r.RequestURI)
			if r.RequestURI != "/api/internal/stats" {
				next.ServeHTTP(w, r)
				return
			}
			// get user ip
			userIP, err := resolveIP(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			// get trusted subnet
			_, subnet, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			// check user ip is located in trusted subnet
			if !subnet.Contains(userIP) {
				http.Error(w, "ip is not trusted", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func resolveIP(r *http.Request) (net.IP, error) {
	// check "X-Real-IP" header
	ipStr := r.Header.Get("X-Real-IP")
	ip := net.ParseIP(ipStr)
	if ip == nil {
		// X-Real-IP is empty then try X-Forwarded-For
		ips := r.Header.Get("X-Forwarded-For")
		ipSplit := strings.Split(ips, ",")
		ipStr = ipSplit[0]
		ip = net.ParseIP(ipStr)
	}
	if ip == nil {
		return nil, fmt.Errorf("failed parse ip from http header")
	}
	return ip, nil
}
