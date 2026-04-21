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
	return NewRouterWithJTI(prefix, logger, db, cfg, nil, nil, registry...)
}

// NewRouterWithJTI builds the HTTP mux like NewRouter but also accepts
// a JTI revocation store and an optional RateLimiter for use by the auth handler.
func NewRouterWithJTI(prefix string, logger *Logger, db DatabaseAdapter, cfg *AppConfig, jtiStore *JTIRevocationStore, rl *RateLimiter, registry ...*SchemaRegistry) *http.ServeMux {
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
	authHandler := newAuthSessionHandler(db, cfg, logger, rl)
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
	rqh := newResourceQueryHandlerOrNil(db, reg, cfg)
	rmh := newResourceMutateHandlerOrNil(db, reg, cfg, jtiStore)
	rsh := newResourceSchemaHandlerOrNil(reg, p)
	mux.HandleFunc(fmt.Sprintf("GET %s/data/", p), func(w http.ResponseWriter, r *http.Request) {
		routeDataRequest(w, r, p, http.MethodGet, rqh, rmh, rsh)
	})
	mux.HandleFunc(fmt.Sprintf("POST %s/data/", p), func(w http.ResponseWriter, r *http.Request) {
		routeDataRequest(w, r, p, http.MethodPost, rqh, rmh, rsh)
	})

	return mux
}

// newResourceQueryHandlerOrNil creates a ResourceQueryHandler if dependencies
// are available, otherwise returns nil.
func newResourceQueryHandlerOrNil(db DatabaseAdapter, reg *SchemaRegistry, cfg *AppConfig) *ResourceQueryHandler {
	if db == nil || reg == nil || cfg == nil {
		return nil
	}
	return NewResourceQueryHandler(db, reg, cfg)
}

// newResourceMutateHandlerOrNil creates a ResourceMutateHandler if dependencies
// are available, otherwise returns nil.
func newResourceMutateHandlerOrNil(db DatabaseAdapter, reg *SchemaRegistry, cfg *AppConfig, jtiStore *JTIRevocationStore) *ResourceMutateHandler {
	if db == nil || reg == nil || cfg == nil {
		return nil
	}
	return NewResourceMutateHandler(db, reg, cfg, jtiStore)
}

// newResourceSchemaHandlerOrNil creates a ResourceSchemaHandler if the
// registry is available, otherwise returns nil.
func newResourceSchemaHandlerOrNil(reg *SchemaRegistry, prefix string) *ResourceSchemaHandler {
	if reg == nil {
		return nil
	}
	return NewResourceSchemaHandler(reg, prefix)
}

// routeDataRequest dispatches /data/{resource}:{action} paths to the
// appropriate handler based on the action suffix.
func routeDataRequest(w http.ResponseWriter, r *http.Request, prefix, method string, rqh *ResourceQueryHandler, rmh *ResourceMutateHandler, rsh *ResourceSchemaHandler) {
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
		if rqh != nil {
			rqh.HandleQuery(w, r)
		} else {
			handleResourceQuery(w, r)
		}
	case method == http.MethodPost && action == "mutate":
		if rmh != nil {
			rmh.HandleMutate(w, r)
		} else {
			handleResourceMutate(w, r)
		}
	case method == http.MethodGet && action == "schema":
		if rsh != nil {
			rsh.HandleSchema(w, r)
		} else {
			handleResourceSchema(w, r)
		}
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
	// Final request order:
	//   method validation → CORS → panic recovery → audit context → auth → website origin → rate limit → captcha → authz → handler
	if bo.authMiddleware != nil {
		handler = Authorize(cfg.Server.Prefix, handler)
		if bo.captchaStore != nil {
			handler = captchaMiddleware(bo.captchaStore, handler)
		}
		if bo.rateLimiter != nil {
			handler = rateLimitMiddleware(bo.rateLimiter, logger, handler)
		}
		handler = websiteAPIKeyMiddleware(handler)
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
	rateLimiter    *RateLimiter
	captchaStore   *CaptchaStore
}

// BuildHandlerOption configures optional BuildHandler dependencies.
type BuildHandlerOption func(*buildHandlerOptions)

// WithAuthMiddleware adds authentication and authorization middleware.
func WithAuthMiddleware(am *AuthMiddleware) BuildHandlerOption {
	return func(o *buildHandlerOptions) {
		o.authMiddleware = am
	}
}

// WithRateLimiter adds rate limiting middleware for authenticated requests.
func WithRateLimiter(rl *RateLimiter) BuildHandlerOption {
	return func(o *buildHandlerOptions) {
		o.rateLimiter = rl
	}
}

// WithCaptchaStore adds CAPTCHA validation for API keys that require it.
func WithCaptchaStore(store *CaptchaStore) BuildHandlerOption {
	return func(o *buildHandlerOptions) {
		o.captchaStore = store
	}
}

// StartServer creates and starts the HTTP server with graceful shutdown.
// It blocks until the server shuts down.
func StartServer(cfg *AppConfig, logger *Logger, db ...DatabaseAdapter) error {
	var adapter DatabaseAdapter
	if len(db) > 0 {
		adapter = db[0]
	}

	var handlerOpts []BuildHandlerOption
	var jtiStore *JTIRevocationStore
	var rl *RateLimiter
	var captchaStore *CaptchaStore
	if adapter != nil && cfg.JWTSecret != "" {
		jtiStore = NewJTIRevocationStore()
		rl = NewRateLimiter()
		captchaStore = NewCaptchaStore()
		am := NewAuthMiddleware(adapter, cfg.JWTSecret, cfg.Server.Prefix, jtiStore)
		handlerOpts = append(handlerOpts, WithAuthMiddleware(am))
		handlerOpts = append(handlerOpts, WithRateLimiter(rl))
		handlerOpts = append(handlerOpts, WithCaptchaStore(captchaStore))
	}

	var reg *SchemaRegistry
	if adapter != nil {
		var err error
		reg, err = NewSchemaRegistry(adapter)
		if err != nil {
			return fmt.Errorf("create schema registry: %w", err)
		}
	}

	mux := NewRouterWithJTI(cfg.Server.Prefix, logger, adapter, cfg, jtiStore, rl, reg)
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
