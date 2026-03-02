package spellcheck

import (
	"errors"
	"fmt"
)

type Config struct {
	Enabled  bool
	MaxItems int
}

func DefaultConfig() Config {
	return Config{
		Enabled:  true,
		MaxItems: 100,
	}
}

func (c *Config) Validate() error {
	if c.MaxItems < 1 {
		return errors.New("max_items must be greater than 0")
	}

	if c.MaxItems > 1000 {
		return fmt.Errorf("max_items cannot exceed 1000, got %d", c.MaxItems)
	}

	return nil
}
