package memory

import (
	"database/sql"
)

// FTSSearch performs full-text search using SQLite FTS5
type FTSSearch struct {
	db *sql.DB
}

// NewFTSSearch creates a new FTS search handler
func NewFTSSearch(db *sql.DB) *FTSSearch {
	return &FTSSearch{db: db}
}

// Search performs a full-text search using FTS5 or falls back to LIKE search
func (fts *FTSSearch) Search(query string, limit int) ([]FTSResult, error) {
	if limit <= 0 {
		limit = 10
	}

	results, err := fts.searchWithFTS5(query, limit)
	if err != nil {
		return fts.searchWithFallback(query, limit)
	}
	return results, nil
}

// searchWithFTS5 attempts FTS5 search
func (fts *FTSSearch) searchWithFTS5(query string, limit int) ([]FTSResult, error) {
	sqlQuery := `
		SELECT 
			m.id,
			m.type,
			m.date,
			m.content,
			m.created_at,
			m.updated_at,
			m.access_count,
			m.last_accessed,
			rank
		FROM memory_fts
		JOIN memory_entries m ON m.id = memory_fts.rowid
		WHERE memory_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`

	rows, err := fts.db.Query(sqlQuery, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return fts.scanResults(rows)
}

// searchWithFallback uses LIKE search when FTS5 is unavailable
func (fts *FTSSearch) searchWithFallback(query string, limit int) ([]FTSResult, error) {
	sqlQuery := `
		SELECT 
			id,
			type,
			date,
			content,
			created_at,
			updated_at,
			access_count,
			last_accessed,
			1.0 as rank
		FROM memory_entries
		WHERE content LIKE ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	pattern := "%" + query + "%"
	rows, err := fts.db.Query(sqlQuery, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return fts.scanResults(rows)
}

// scanResults scans query results into FTSResult slice
func (fts *FTSSearch) scanResults(rows *sql.Rows) ([]FTSResult, error) {
	var results []FTSResult
	for rows.Next() {
		var result FTSResult
		var lastAccessed sql.NullTime
		err := rows.Scan(
			&result.Entry.ID,
			&result.Entry.Type,
			&result.Entry.Date,
			&result.Entry.Content,
			&result.Entry.CreatedAt,
			&result.Entry.UpdatedAt,
			&result.Entry.AccessCount,
			&lastAccessed,
			&result.Rank,
		)
		if err != nil {
			continue
		}
		if lastAccessed.Valid {
			result.Entry.LastAccessed = &lastAccessed.Time
		}
		results = append(results, result)
	}
	return results, rows.Err()
}

// SearchWithFilter performs search with additional filters
func (fts *FTSSearch) SearchWithFilter(query string, opts SearchOptions) ([]FTSResult, error) {
	if opts.Limit <= 0 {
		opts.Limit = 10
	}

	results, err := fts.searchWithFilterFTS5(query, opts)
	if err != nil {
		return fts.searchWithFilterFallback(query, opts)
	}
	return results, nil
}

func (fts *FTSSearch) searchWithFilterFTS5(query string, opts SearchOptions) ([]FTSResult, error) {
	baseQuery := `
		SELECT 
			m.id,
			m.type,
			m.date,
			m.content,
			m.created_at,
			m.updated_at,
			m.access_count,
			m.last_accessed,
			rank
		FROM memory_fts
		JOIN memory_entries m ON m.id = memory_fts.rowid
		WHERE memory_fts MATCH ?
	`

	args := []interface{}{query}

	if opts.Type != "" {
		baseQuery += " AND m.type = ?"
		args = append(args, opts.Type)
	}

	if opts.Date != "" {
		baseQuery += " AND m.date = ?"
		args = append(args, opts.Date)
	}

	baseQuery += " ORDER BY rank LIMIT ?"
	args = append(args, opts.Limit)

	rows, err := fts.db.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return fts.scanResults(rows)
}

func (fts *FTSSearch) searchWithFilterFallback(query string, opts SearchOptions) ([]FTSResult, error) {
	baseQuery := `
		SELECT 
			id,
			type,
			date,
			content,
			created_at,
			updated_at,
			access_count,
			last_accessed,
			1.0 as rank
		FROM memory_entries
		WHERE content LIKE ?
	`

	args := []interface{}{"%" + query + "%"}

	if opts.Type != "" {
		baseQuery += " AND type = ?"
		args = append(args, opts.Type)
	}

	if opts.Date != "" {
		baseQuery += " AND date = ?"
		args = append(args, opts.Date)
	}

	baseQuery += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, opts.Limit)

	rows, err := fts.db.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return fts.scanResults(rows)
}

// Insert adds a new entry to the FTS index
func (fts *FTSSearch) Insert(entryID string, content string) error {
	// FTS5 virtual table is automatically populated via content_rowid
	// when we insert into memory_entries
	return nil
}

// Delete removes an entry from the FTS index
func (fts *FTSSearch) Delete(entryID string) error {
	// FTS5 is automatically maintained via foreign key
	return nil
}

// Rebuild rebuilds the FTS index (useful for maintenance)
func (fts *FTSSearch) Rebuild() error {
	_, err := fts.db.Exec("INSERT INTO memory_fts(memory_fts) VALUES('rebuild')")
	return err
}

// Optimize runs FTS5 optimization (should be called periodically)
func (fts *FTSSearch) Optimize() error {
	_, err := fts.db.Exec("INSERT INTO memory_fts(memory_fts) VALUES('optimize')")
	return err
}

// FTSResult represents a single FTS search result
type FTSResult struct {
	Entry MemoryEntry
	Rank  float64
}
