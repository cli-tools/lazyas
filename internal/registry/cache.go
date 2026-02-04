package registry

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
	"lazyas/internal/config"
)

// Cache represents the cached index
type Cache struct {
	Index     *Index    `yaml:"index"`
	FetchedAt time.Time `yaml:"fetched_at"`
}

// CacheManager handles index caching
type CacheManager struct {
	cfg   *config.Config
	cache *Cache
}

// NewCacheManager creates a new cache manager
func NewCacheManager(cfg *config.Config) *CacheManager {
	return &CacheManager{
		cfg: cfg,
	}
}

// Load reads the cache from disk
func (c *CacheManager) Load() error {
	data, err := os.ReadFile(c.cfg.CachePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.cache = nil
			return nil
		}
		return err
	}

	var cache Cache
	if err := yaml.Unmarshal(data, &cache); err != nil {
		return err
	}

	c.cache = &cache
	return nil
}

// Save writes the cache to disk
func (c *CacheManager) Save() error {
	if c.cache == nil {
		return nil
	}

	if err := c.cfg.EnsureDirs(); err != nil {
		return err
	}

	data, err := yaml.Marshal(c.cache)
	if err != nil {
		return err
	}

	return os.WriteFile(c.cfg.CachePath, data, 0644)
}

// IsValid checks if the cache is valid
func (c *CacheManager) IsValid() bool {
	if c.cache == nil || c.cache.Index == nil {
		return false
	}

	ttl := time.Duration(c.cfg.CacheTTL) * time.Hour
	return time.Since(c.cache.FetchedAt) < ttl
}

// Get returns the cached index
func (c *CacheManager) Get() *Index {
	if c.cache == nil {
		return nil
	}
	return c.cache.Index
}

// Set updates the cache
func (c *CacheManager) Set(index *Index) error {
	c.cache = &Cache{
		Index:     index,
		FetchedAt: time.Now(),
	}
	return c.Save()
}
