package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	bucketProcessed = []byte("processed")
	bucketPrefs     = []byte("userprefs")
	bucketSudoku    = []byte("sudoku")
)

// Store persists processed message IDs across restarts using BoltDB.
type Store struct {
	db *bolt.DB
}

type record struct {
	ProcessedAt time.Time `json:"processed_at"`
}

// UserPrefs holds per-user settings persisted across restarts.
type UserPrefs struct {
	AIModel string `json:"ai_model,omitempty"`
}

func New(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return nil, fmt.Errorf("store: create dir: %w", err)
	}
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("store: open db: %w", err)
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{bucketProcessed, bucketPrefs, bucketSudoku} {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: init bucket: %w", err)
	}
	return &Store{db: db}, nil
}

// IsProcessed reports whether the message with the given ID has already been handled.
func (s *Store) IsProcessed(id string) (bool, error) {
	var found bool
	err := s.db.View(func(tx *bolt.Tx) error {
		found = tx.Bucket(bucketProcessed).Get([]byte(id)) != nil
		return nil
	})
	return found, err
}

// MarkProcessed records the message ID as handled.
func (s *Store) MarkProcessed(id string) error {
	data, _ := json.Marshal(record{ProcessedAt: time.Now().UTC()})
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketProcessed).Put([]byte(id), data)
	})
}

// GetUserPrefs retrieves saved preferences for the given email address.
// Returns empty prefs (no error) if nothing is stored yet.
func (s *Store) GetUserPrefs(email string) (UserPrefs, error) {
	var prefs UserPrefs
	err := s.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(bucketPrefs).Get([]byte(email))
		if data == nil {
			return nil
		}
		return json.Unmarshal(data, &prefs)
	})
	return prefs, err
}

// SetUserPrefs persists preferences for the given email address.
func (s *Store) SetUserPrefs(email string, prefs UserPrefs) error {
	data, err := json.Marshal(prefs)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketPrefs).Put([]byte(email), data)
	})
}

// ── Sudoku puzzle storage (2-day expiry) ─────────────────────────────────────

// SudokuEntry holds a puzzle and its solution with an expiry timestamp.
type SudokuEntry struct {
	ID         string    `json:"id"`
	Puzzle     string    `json:"puzzle"`
	Solution   string    `json:"solution"`
	Difficulty string    `json:"difficulty"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// SaveSudoku persists a puzzle entry.
func (s *Store) SaveSudoku(e SudokuEntry) error {
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketSudoku).Put([]byte(e.ID), data)
	})
}

// GetSudoku looks up a puzzle by ID. Returns nil (no error) if not found or expired.
func (s *Store) GetSudoku(id string) (*SudokuEntry, error) {
	var entry SudokuEntry
	var found bool

	if err := s.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(bucketSudoku).Get([]byte(id))
		if data == nil {
			return nil
		}
		found = true
		return json.Unmarshal(data, &entry)
	}); err != nil || !found {
		return nil, err
	}

	if time.Now().After(entry.ExpiresAt) {
		_ = s.db.Update(func(tx *bolt.Tx) error {
			return tx.Bucket(bucketSudoku).Delete([]byte(id))
		})
		return nil, nil
	}
	return &entry, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
