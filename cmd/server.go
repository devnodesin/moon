package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// NewRouter builds the HTTP mux with all routes registered under the
// configured server prefix.
func NewRouter(prefix string, logger *Logger, db DatabaseAdapter, cfg *AppConfig, registry ...*SchemaRegistry) *http.ServeMux {
	mux := http.NewServeMux()

	p := strings.TrimRight(prefix, "/")

	// Public routes
	mux.HandleFunc(fmt.Sprintf("GET %s/health", p), handleHealth)
	if p == "" {
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				WriteError(w, http.StatusNotFound, "Not found")
				return
			}
			handleHealth(w, r)
		})
	} else {
		// With prefix: GET /prefix → health, GET /prefix/ → health
		mux.HandleFunc(fmt.Sprintf("GET %s", p), handleHealth)
		mux.HandleFunc(fmt.Sprintf("GET %s/", p), func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != p+"/" {
				WriteError(w, http.StatusNotFound, "Not found")
				return
			}
			handleHealth(w, r)
		})
	}

	// Auth routes
	authHandler := newAuthSessionHandler(db, cfg)
	mux.HandleFunc(fmt.Sprintf("POST %s/auth:session", p), authHandler.HandleSession)

	authMeHandler := NewAuthMeHandler(db, cfg)
	mux.HandleFunc(fmt.Sprintf("GET %s/auth:me", p), authMeHandler.GetMe)
	mux.HandleFunc(fmt.Sprintf("POST %s/auth:me", p), authMeHandler.UpdateMe)

	// Collection routes
	var reg *SchemaRegistry
	if len(registry) > 0 {
		reg = registry[0]
	}
	if reg != nil && db != nil {
		ch := NewCollectionHandler(db, reg, cfg)
		mux.HandleFunc(fmt.Sprintf("GET %s/collections:query", p), ch.HandleQuery)
		mux.HandleFunc(fmt.Sprintf("POST %s/collections:mutate", p), ch.HandleMutate)
	} else {
		mux.HandleFunc(fmt.Sprintf("GET %s/collections:query", p), handleCollectionsQuery)
		mux.HandleFunc(fmt.Sprintf("POST %s/collections:mutate", p), handleCollectionsMutate)
	}

	// Resource routes — use a catch-all pattern for /data/ paths
	mux.HandleFunc(fmt.Sprintf("GET %s/data/", p), func(w http.ResponseWriter, r *http.Request) {
		routeDataRequest(w, r, p, http.MethodGet)
	})
	mux.HandleFunc(fmt.Sprintf("POST %s/data/", p), func(w http.ResponseWriter, r *http.Request) {
		routeDataRequest(w, r, p, http.MethodPost)
	})

	return mux
}

// routeDataRequest dispatches /data/{resource}:{action} paths to the
// appropriate handler based on the action suffix.
func routeDataRequest(w http.ResponseWriter, r *http.Request, prefix, method string) {
	path := r.URL.Path
	dataPrefix := prefix + "/data/"
	if !strings.HasPrefix(path, dataPrefix) {
		WriteError(w, http.StatusNotFound, "Not found")
		return
	}

	rest := path[len(dataPrefix):]

	colonIdx := strings.LastIndex(rest, ":")
	if colonIdx < 0 {
		WriteError(w, http.StatusNotFound, "Not found")
		return
	}

	resource := rest[:colonIdx]
	action := rest[colonIdx+1:]

	if resource == "" {
		WriteError(w, http.StatusBadRequest, "Missing resource name")
		return
	}

	if strings.HasPrefix(resource, "moon_") {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("Resource name %q is reserved", resource))
		return
	}

	switch {
	case method == http.MethodGet && action == "query":
		handleResourceQuery(w, r)
	case method == http.MethodPost && action == "mutate":
		handleResourceMutate(w, r)
	case method == http.MethodGet && action == "schema":
		handleResourceSchema(w, r)
	default:
		WriteError(w, http.StatusNotFound, "Not found")
	}
}

// BuildHandler wraps the router with the full middleware chain in the order
// specified by SPEC.md §6.2.
func BuildHandler(mux *http.ServeMux, cfg *AppConfig, logger *Logger, opts ...BuildHandlerOption) http.Handler {
	var bo buildHandlerOptions
	for _, o := range opts {
		o(&bo)
	}

	var handler http.Handler = mux

	// Middleware wraps from inside out, so we apply in reverse order.
	// Final order: method validation → CORS → panic recovery → audit context → auth → authz → handler
	if bo.authMiddleware != nil {
		handler = Authorize(cfg.Server.Prefix, handler)
		handler = bo.authMiddleware.Authenticate(handler)
	}
	handler = auditContextMiddleware(logger, handler)
	handler = panicRecoveryMiddleware(logger, handler)
	handler = corsMiddleware(cfg.CORS, handler)
	handler = methodValidationMiddleware(handler)

	return handler
}

// buildHandlerOptions holds optional dependencies for BuildHandler.
type buildHandlerOptions struct {
	authMiddleware *AuthMiddleware
}

// BuildHandlerOption configures optional BuildHandler dependencies.
type BuildHandlerOption func(*buildHandlerOptions)

// WithAuthMiddleware adds authentication and authorization middleware.
func WithAuthMiddleware(am *AuthMiddleware) BuildHandlerOption {
	return func(o *buildHandlerOptions) {
		o.authMiddleware = am
	}
}

// StartServer creates and starts the HTTP server with graceful shutdown.
// It blocks until the server shuts down.
func StartServer(cfg *AppConfig, logger *Logger, db ...DatabaseAdapter) error {
	var adapter DatabaseAdapter
	if len(db) > 0 {
		adapter = db[0]
	}
	mux := NewRouter(cfg.Server.Prefix, logger, adapter, cfg)

	var handlerOpts []BuildHandlerOption
	if adapter != nil && cfg.JWTSecret != "" {
		jtiStore := NewJTIRevocationStore()
		am := NewAuthMiddleware(adapter, cfg.JWTSecret, cfg.Server.Prefix, jtiStore)
		handlerOpts = append(handlerOpts, WithAuthMiddleware(am))
	}
	handler := BuildHandler(mux, cfg, logger, handlerOpts...)

	addr := net.JoinHostPort(cfg.Server.Host, fmt.Sprintf("%d", cfg.Server.Port))
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", addr, "prefix", cfg.Server.Prefix)
		logger.AuditEvent(AuditStartupSuccess, "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	select {
	case err := <-errCh:
		return fmt.Errorf("server failed: %w", err)
	case sig := <-sigCh:
		logger.Info("shutdown signal received", "signal", sig.String())
		logger.AuditEvent(AuditShutdown, "reason", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	logger.Info("server stopped gracefully")
	return nil
}
