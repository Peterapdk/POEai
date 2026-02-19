package memory

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

type MemoryType string

const (
	TypeEpisodic   MemoryType = "episodic"
	TypeSemantic   MemoryType = "semantic"
	TypeFact       MemoryType = "fact"
	TypeProcedural MemoryType = "procedural"
)

type Memory struct {
	ID         string                 `json:"id"`
	Type       MemoryType             `json:"type"`
	Content    string                 `json:"content"`
	Source     string                 `json:"source"`
	Importance float64                `json:"importance"`
	CreatedAt  time.Time              `json:"created_at"`
	AccessedAt time.Time              `json:"accessed_at"`
	Metadata   map[string]interface{} `json:"metadata"`
}

type Store struct {
	db *sql.DB
}

// Open opens a SQLite database at the given path.
func Open(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if _, err := db.Exec(schemaSQL); err != nil {
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Write(ctx context.Context, mem Memory) (string, error) {
	if mem.ID == "" {
		mem.ID = uuid.New().String()
	}
	if mem.CreatedAt.IsZero() {
		mem.CreatedAt = time.Now()
	}
	if mem.AccessedAt.IsZero() {
		mem.AccessedAt = time.Now()
	}
	if mem.Importance == 0 {
		mem.Importance = 0.5
	}
	if mem.Metadata == nil {
		mem.Metadata = make(map[string]interface{})
	}

	metaJSON, _ := json.Marshal(mem.Metadata)

	query := `INSERT INTO memories (id, type, content, source, importance, created_at, accessed_at, metadata)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query,
		mem.ID, string(mem.Type), mem.Content, mem.Source, mem.Importance,
		mem.CreatedAt.Unix(), mem.AccessedAt.Unix(), string(metaJSON))
	if err != nil {
		return "", fmt.Errorf("insert memory: %w", err)
	}

	return mem.ID, nil
}

func (s *Store) Search(ctx context.Context, query string, limit int) ([]Memory, error) {
	// Simple LIKE search for v1 (Task 2). Task 5 will add sqlite-vec.
	sqlQuery := `SELECT id, type, content, source, importance, created_at, accessed_at, metadata
	             FROM memories
	             WHERE content LIKE ?
	             ORDER BY importance DESC, accessed_at DESC
	             LIMIT ?`

	rows, err := s.db.QueryContext(ctx, sqlQuery, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer rows.Close()

	var results []Memory
	for rows.Next() {
		var m Memory
		var mType, metaStr string
		var created, accessed int64
		err := rows.Scan(&m.ID, &mType, &m.Content, &m.Source, &m.Importance, &created, &accessed, &metaStr)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		m.Type = MemoryType(mType)
		m.CreatedAt = time.Unix(created, 0)
		m.AccessedAt = time.Unix(accessed, 0)
		json.Unmarshal([]byte(metaStr), &m.Metadata)
		results = append(results, m)
	}

	return results, nil
}

func (s *Store) SetFact(ctx context.Context, key, value string, confidence float64) error {
	query := `INSERT INTO facts (key, value, confidence, updated_at) 
	          VALUES (?, ?, ?, ?)
	          ON CONFLICT(key) DO UPDATE SET value=excluded.value, confidence=excluded.confidence, updated_at=excluded.updated_at`
	_, err := s.db.ExecContext(ctx, query, key, value, confidence, time.Now().Unix())
	return err
}

func (s *Store) GetFact(ctx context.Context, key string) (string, bool, error) {
	var val string
	err := s.db.QueryRowContext(ctx, "SELECT value FROM facts WHERE key = ?", key).Scan(&val)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}
