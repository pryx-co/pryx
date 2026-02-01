# RAG Memory System

Pryx includes a sophisticated RAG (Retrieval-Augmented Generation) memory system for long-term knowledge retention. This system enables the AI agent to store, retrieve, and search across past conversations and curated knowledge.

## Architecture Overview

The memory system is built on a three-layer architecture:

### 1. Daily Logs (`memory/YYYY-MM-DD`)
Append-only daily context logs that capture the natural flow of conversations and work. These are automatically created and represent the raw history of agent interactions.

**Use Cases:**
- Recording session summaries
- Tracking daily work context
- Capturing important decisions

### 2. Long-term Memory (`MEMORY`)
Curated persistent knowledge that the agent explicitly decides to preserve. This is high-value information that should survive across sessions.

**Use Cases:**
- Project specifications and requirements
- Learned patterns and insights
- Important user preferences
- Critical conclusions from analysis

### 3. Session Memory
Current conversation context that exists in the session. This layer is ephemeral and exists primarily in the message history.

## Storage Backend

The memory system uses **SQLite** for robust, ACID-compliant storage with the following extensions:

- **FTS5** (Full-Text Search) - for keyword-based search with ranking
- **sqlite-vec** (future) - for vector similarity search when embedding providers are added

### Database Schema

```sql
-- Core memory entries
CREATE TABLE memory_entries (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,        -- 'daily', 'longterm', 'session'
    date TEXT,                 -- YYYY-MM-DD for daily logs
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    access_count INTEGER DEFAULT 0,
    last_accessed DATETIME
);

-- Full-text search index
CREATE VIRTUAL TABLE memory_fts USING fts5(
    content,
    content_rowid=id,
    tokenize='porter'
);

-- Source tracking (files, tools, conversations)
CREATE TABLE memory_sources (
    id TEXT PRIMARY KEY,
    entry_id TEXT,
    source_type TEXT,          -- 'file', 'tool', 'conversation'
    source_path TEXT,
    FOREIGN KEY (entry_id) REFERENCES memory_entries(id) ON DELETE CASCADE
);

-- Vector embeddings (placeholder for future)
CREATE TABLE memory_vectors (
    entry_id TEXT PRIMARY KEY,
    embedding BLOB,
    FOREIGN KEY (entry_id) REFERENCES memory_entries(id) ON DELETE CASCADE
);
```

## Search Strategy

The system uses **hybrid search** that combines multiple approaches:

### Full-Text Search (FTS5)
Uses SQLite's FTS5 extension with Porter stemming for keyword matching. Results are ranked by BM25 relevance.

### Vector Search (Placeholder)
Reserved for future implementation when embedding provider plugins are available. The schema supports this but currently returns empty results.

### Hybrid Scoring
Results are combined using weighted scoring:
```
hybrid_score = (fts_score * 0.7) + (vector_score * 0.3)
```

This approach prioritizes keyword matching while leaving room for semantic similarity when embeddings are available.

## Auto-Flush System

The memory system includes an auto-flush mechanism that triggers before context window compaction:

1. **Monitoring**: Tracks token count approaching the context limit
2. **Trigger**: When tokens approach the threshold (default: 80% of limit)
3. **Action**: Silently reminds the agent to store durable memories
4. **Storage**: Key learnings are written to long-term memory

### Configuration

```yaml
# config.yaml
memory_enabled: true              # Enable RAG memory system
memory_auto_flush: true           # Enable auto-flush before compaction
memory_flush_threshold_tokens: 100000  # Token threshold for flush trigger
```

## API Endpoints

The memory system exposes REST API endpoints:

### List Memory Entries
```
GET /api/v1/memory?type=daily&limit=50
```

Query Parameters:
- `type`: Filter by type ('daily', 'longterm', 'session')
- `date`: Filter by date (YYYY-MM-DD format)
- `limit`: Maximum results (default: 100)

### Write Memory
```
POST /api/v1/memory
Content-Type: application/json

{
  "type": "longterm",
  "content": "Important insight learned from analysis...",
  "sources": [
    {"source_type": "conversation", "source_path": "session_123"}
  ]
}
```

### Search Memory
```
POST /api/v1/memory/search
Content-Type: application/json

{
  "query": "authentication implementation",
  "type": "longterm",
  "limit": 10,
  "include_fts": true,
  "include_vector": false
}
```

## Agent Integration

The memory system is automatically integrated into the agent's system prompt. When enabled, the agent receives:

1. **Relevant context** from memory search results
2. **Instructions** on how to use the memory system
3. **Reminders** to store important learnings

### System Prompt Addition

```
=== RELEVANT MEMORY ===

[1] longterm (2026-01-15):
Important authentication pattern discovered: OAuth2 with PKCE
is the recommended approach for mobile applications...

[2] daily (2026-01-20):
Session summary: Completed refactoring of the auth module...
```

## Sovereign-First Design

The RAG memory system is designed with sovereignty as a core principle:

### No External Dependencies
- **v1.0**: FTS5 search only, no external embedding providers
- **Future**: Embedding provider plugins can be added as optional extensions

### Local Storage
- All data stored in local SQLite database
- No cloud synchronization by default
- Full data ownership by the user

### Extensibility
The schema is designed to support future embedding providers:
- `memory_vectors` table ready for vector storage
- Hybrid search weights can be adjusted
- Provider plugins can populate embeddings

## Future Extensions

### Embedding Provider Plugins
When embedding providers are added, they will:
1. Generate embeddings for new memory entries
2. Store vectors in `memory_vectors` table
3. Enable semantic similarity search
4. Work alongside existing FTS5 search

### Planned Providers
- Local models (Ollama, etc.)
- OpenAI (optional, user-configured)
- Anthropic (optional, user-configured)
- Custom provider interface

## Usage Examples

### Writing Daily Log
```go
sources := []memory.MemorySource{
    {SourceType: "conversation", SourcePath: sessionID},
}
entryID, err := ragManager.WriteDaily("Completed OAuth implementation", sources)
```

### Writing Long-term Memory
```go
sources := []memory.MemorySource{
    {SourceType: "file", SourcePath: "/docs/auth.md"},
}
entryID, err := ragManager.WriteLongterm("OAuth2 with PKCE is recommended for mobile apps", sources)
```

### Searching Memory
```go
results, err := ragManager.Search(ctx, "authentication best practices", memory.SearchOptions{
    Type:       memory.MemoryTypeLongterm,
    Limit:      5,
    IncludeFTS: true,
})
```

### Auto-Flush Check
```go
if ragMemory.AutoFlush().ShouldFlush(tokenCount, maxTokens, thresholdTokens) {
    reminder := ragMemory.AutoFlush().GetFlushReminder()
    // Include reminder in system prompt
}
```

## Performance Considerations

- **FTS5 Maintenance**: Run `OPTIMIZE` periodically to maintain index performance
- **Access Tracking**: Access counts are updated on each retrieval for analytics
- **Pagination**: Use `limit` parameter for large result sets
- **Filtering**: Filter by type and date to narrow search scope

## Security

- Memory entries are stored in the same SQLite database as sessions
- No encryption by default (database-level encryption can be added)
- Source tracking maintains provenance of information
