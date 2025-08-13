/*
 * Copyright (C) 2025 Miguel Mamani <miguel.coder.per@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>.
 */

// cmd/api/main.go
package main

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"highperf-api/internal/httpserver"
)

func main() {
	router := httpserver.NewRouter()
	srv := &http.Server{
		Handler:           router,
		ReadTimeout:       2 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		WriteTimeout:      2 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    8 << 10, // 8KB
		// TLSConfig: opcional si vas con TLS directo (en prod suele ir detrás de LB)
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			// PreferServerCipherSuites se ignora en TLS1.3, está ok en 1.2
			PreferServerCipherSuites: true,
		},
	}

	// SO_REUSEPORT para escalar por proceso (Linux)
	ln, err := reusePortListen("tcp", ":8080")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	// Arranque del servidor
	go func() {
		log.Printf("listening on %s", ln.Addr())
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatalf("serve: %v", err)
		}
	}()

	// Apagado elegante
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
}

func reusePortListen(network, address string) (net.Listener, error) {
	// Usa un lib probado en producción:
	// github.com/libp2p/go-reuseport o github.com/kavu/go_reuseport
	// Aquí lo dejamos simple para mantener el snippet autocontenido:
	return net.Listen(network, address)
}
