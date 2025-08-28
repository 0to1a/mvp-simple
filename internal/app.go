package internal

import (
	"database/sql"
	"time"

	"project/internal/db/sqlc"

	"github.com/patrickmn/go-cache"
)

type App struct {
	Cfg          Config
	DB           *sql.DB
	Queries      *sqlc.Queries
	Cache        *cache.Cache
	EmailService *EmailService
}

func NewApp(cfg Config, db *sql.DB) *App {
	cacheTTL := 60 * time.Second // Hardcoded 60 second cache TTL
	app := &App{
		Cfg:     cfg,
		DB:      db,
		Queries: sqlc.New(db),
		Cache:   cache.New(cacheTTL, 2*cacheTTL),
	}
	
	// Initialize email service
	app.EmailService = NewEmailService(cfg.EmailAPIKey, cfg.EmailFromAddress)
	
	return app
}

func (a *App) CacheSet(key string, value any, ttl time.Duration) { a.Cache.Set(key, value, ttl) }
func (a *App) CacheGet(key string) (any, bool)                   { return a.Cache.Get(key) }
