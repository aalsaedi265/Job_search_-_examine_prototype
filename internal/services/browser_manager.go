package services

import (
	"context"
	"log"
	"sync"
	"time"
)

// BrowserContext holds a browser context and its metadata
type BrowserContext struct {
	Ctx       context.Context
	Cancel    context.CancelFunc
	CreatedAt time.Time
	AllocCtx  context.Context
	AllocCancel context.CancelFunc
}

// BrowserManager manages active browser contexts for paused applications
type BrowserManager struct {
	contexts map[string]*BrowserContext // application_id -> context
	mu       sync.RWMutex
	timeout  time.Duration
}

// NewBrowserManager creates a new browser manager with the specified timeout
func NewBrowserManager(timeout time.Duration) *BrowserManager {
	bm := &BrowserManager{
		contexts: make(map[string]*BrowserContext),
		timeout:  timeout,
	}

	// Start cleanup ticker
	go bm.cleanupLoop()

	return bm
}

// Store saves a browser context for an application
func (bm *BrowserManager) Store(applicationID string, ctx context.Context, cancel context.CancelFunc, allocCtx context.Context, allocCancel context.CancelFunc) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Clean up old context if exists
	if old, exists := bm.contexts[applicationID]; exists {
		log.Printf("Replacing existing browser context for application %s", applicationID)
		old.Cancel()
		if old.AllocCancel != nil {
			old.AllocCancel()
		}
	}

	bm.contexts[applicationID] = &BrowserContext{
		Ctx:         ctx,
		Cancel:      cancel,
		CreatedAt:   time.Now(),
		AllocCtx:    allocCtx,
		AllocCancel: allocCancel,
	}

	log.Printf("Stored browser context for application %s (total active: %d)", applicationID, len(bm.contexts))
}

// Get retrieves a browser context for an application
func (bm *BrowserManager) Get(applicationID string) (*BrowserContext, bool) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	ctx, exists := bm.contexts[applicationID]
	return ctx, exists
}

// Remove removes and cancels a browser context
func (bm *BrowserManager) Remove(applicationID string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if ctx, exists := bm.contexts[applicationID]; exists {
		log.Printf("Removing browser context for application %s", applicationID)
		ctx.Cancel()
		if ctx.AllocCancel != nil {
			ctx.AllocCancel()
		}
		delete(bm.contexts, applicationID)
	}
}

// CleanupExpired removes contexts that have exceeded the timeout
func (bm *BrowserManager) CleanupExpired() int {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	now := time.Now()
	expired := []string{}

	for appID, ctx := range bm.contexts {
		if now.Sub(ctx.CreatedAt) > bm.timeout {
			expired = append(expired, appID)
		}
	}

	for _, appID := range expired {
		log.Printf("Cleaning up expired browser context for application %s (age: %s)",
			appID, now.Sub(bm.contexts[appID].CreatedAt))
		bm.contexts[appID].Cancel()
		if bm.contexts[appID].AllocCancel != nil {
			bm.contexts[appID].AllocCancel()
		}
		delete(bm.contexts, appID)
	}

	if len(expired) > 0 {
		log.Printf("Cleaned up %d expired browser contexts (remaining: %d)", len(expired), len(bm.contexts))
	}

	return len(expired)
}

// Count returns the number of active contexts
func (bm *BrowserManager) Count() int {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return len(bm.contexts)
}

// cleanupLoop runs every minute to clean up expired contexts
func (bm *BrowserManager) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		bm.CleanupExpired()
	}
}

// Shutdown cancels all active contexts
func (bm *BrowserManager) Shutdown() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	log.Printf("Shutting down browser manager, canceling %d active contexts", len(bm.contexts))

	for appID, ctx := range bm.contexts {
		log.Printf("Canceling browser context for application %s", appID)
		ctx.Cancel()
		if ctx.AllocCancel != nil {
			ctx.AllocCancel()
		}
	}

	bm.contexts = make(map[string]*BrowserContext)
}
