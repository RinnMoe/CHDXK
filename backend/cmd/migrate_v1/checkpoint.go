package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type migrationCheckpoint struct {
	Path      string                     `json:"-"`
	Stages    map[string]checkpointStage `json:"stages"`
	UpdatedAt time.Time                  `json:"updated_at"`
}

type checkpointStage struct {
	LastID    int       `json:"last_id"`
	Done      bool      `json:"done"`
	UpdatedAt time.Time `json:"updated_at"`
}

func loadCheckpoint(path string) (*migrationCheckpoint, error) {
	checkpoint := &migrationCheckpoint{
		Path:   path,
		Stages: make(map[string]checkpointStage),
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return checkpoint, nil
		}
		return nil, fmt.Errorf("read checkpoint: %w", err)
	}
	if err := json.Unmarshal(data, checkpoint); err != nil {
		return nil, fmt.Errorf("parse checkpoint: %w", err)
	}
	checkpoint.Path = path
	if checkpoint.Stages == nil {
		checkpoint.Stages = make(map[string]checkpointStage)
	}
	return checkpoint, nil
}

func (c *migrationCheckpoint) StageLastID(name string) int {
	stage := c.Stages[name]
	return stage.LastID
}

func (c *migrationCheckpoint) StageDone(name string) bool {
	stage := c.Stages[name]
	return stage.Done
}

func (c *migrationCheckpoint) MarkProgress(name string, lastID int) error {
	return c.updateStage(name, checkpointStage{LastID: lastID})
}

func (c *migrationCheckpoint) MarkDone(name string) error {
	stage := c.Stages[name]
	stage.Done = true
	return c.updateStage(name, stage)
}

func (c *migrationCheckpoint) updateStage(name string, stage checkpointStage) error {
	now := time.Now()
	stage.UpdatedAt = now
	c.Stages[name] = stage
	c.UpdatedAt = now
	return c.Save()
}

func (c *migrationCheckpoint) Save() error {
	if err := os.MkdirAll(filepath.Dir(c.Path), 0o755); err != nil {
		return fmt.Errorf("create checkpoint directory: %w", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal checkpoint: %w", err)
	}
	tmp := c.Path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write checkpoint temp file: %w", err)
	}
	if err := os.Rename(tmp, c.Path); err != nil {
		return fmt.Errorf("replace checkpoint file: %w", err)
	}
	return nil
}
