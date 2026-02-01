package memory

import (
	"context"
	"fmt"
	"sort"
)

// Search performs hybrid search (FTS5 + placeholder for vector search)
func (m *RAGManager) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	if !m.enabled {
		return nil, fmt.Errorf("memory system is disabled")
	}

	if opts.IncludeFTS {
		return m.hybridSearch(query, opts)
	}

	return m.vectorSearch(query, opts)
}

// hybridSearch combines FTS5 and vector search results
func (m *RAGManager) hybridSearch(query string, opts SearchOptions) ([]SearchResult, error) {
	var results []SearchResult

	ftsResults, err := m.fts.SearchWithFilter(query, opts)
	if err != nil {
		return nil, err
	}

	for _, fts := range ftsResults {
		entry := fts.Entry

		entry.Sources, _ = m.getSources(entry.ID)

		results = append(results, SearchResult{
			Entry:       entry,
			FTSScore:    normalizeScore(fts.Rank),
			VectorScore: 0,
			HybridScore: normalizeScore(fts.Rank),
		})
	}

	if opts.IncludeVector {
		vectorResults, _ := m.vectorSearch(query, opts)
		results = mergeResults(results, vectorResults)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].HybridScore > results[j].HybridScore
	})

	if len(results) > opts.Limit && opts.Limit > 0 {
		results = results[:opts.Limit]
	}

	m.updateAccessCounts(results)

	return results, nil
}

// vectorSearch is a placeholder for future vector similarity search
func (m *RAGManager) vectorSearch(query string, opts SearchOptions) ([]SearchResult, error) {
	return []SearchResult{}, nil
}

// normalizeScore converts FTS rank to 0-1 score
func normalizeScore(rank float64) float64 {
	if rank == 0 {
		return 0
	}
	return 1.0 / (1.0 + rank)
}

// mergeResults combines FTS and vector results with hybrid scoring
func mergeResults(ftsResults, vectorResults []SearchResult) []SearchResult {
	ftsWeight := 0.7
	vectorWeight := 0.3

	resultMap := make(map[string]SearchResult)

	for _, r := range ftsResults {
		resultMap[r.Entry.ID] = r
	}

	for _, r := range vectorResults {
		if existing, ok := resultMap[r.Entry.ID]; ok {
			hybrid := (existing.FTSScore * ftsWeight) + (r.VectorScore * vectorWeight)
			existing.VectorScore = r.VectorScore
			existing.HybridScore = hybrid
			resultMap[r.Entry.ID] = existing
		} else {
			r.HybridScore = r.VectorScore * vectorWeight
			resultMap[r.Entry.ID] = r
		}
	}

	var merged []SearchResult
	for _, r := range resultMap {
		merged = append(merged, r)
	}

	return merged
}

// getSources retrieves sources for a memory entry
func (m *RAGManager) getSources(entryID string) ([]MemorySource, error) {
	rows, err := m.db.Query(
		"SELECT id, source_type, source_path FROM memory_sources WHERE entry_id = ?",
		entryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []MemorySource
	for rows.Next() {
		var s MemorySource
		s.EntryID = entryID
		if err := rows.Scan(&s.ID, &s.SourceType, &s.SourcePath); err == nil {
			sources = append(sources, s)
		}
	}

	return sources, rows.Err()
}

// updateAccessCounts increments access count for retrieved entries
func (m *RAGManager) updateAccessCounts(results []SearchResult) {
	for _, r := range results {
		_, _ = m.db.Exec(
			"UPDATE memory_entries SET access_count = access_count + 1, last_accessed = CURRENT_TIMESTAMP WHERE id = ?",
			r.Entry.ID,
		)
	}
}
