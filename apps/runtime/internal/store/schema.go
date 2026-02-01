package store

const schema = `
CREATE TABLE IF NOT EXISTS sessions (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS messages (
	id TEXT PRIMARY KEY,
	session_id TEXT NOT NULL,
	role TEXT NOT NULL,
	content TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS audit_log (
	id TEXT PRIMARY KEY,
	timestamp DATETIME NOT NULL,
	session_id TEXT,
	surface TEXT,
	tool TEXT,
	action TEXT NOT NULL,
	description TEXT,
	payload TEXT,
	cost TEXT,
	duration INTEGER,
	user_id TEXT,
	success BOOLEAN NOT NULL DEFAULT 1,
	error_msg TEXT,
	metadata TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id);
CREATE INDEX IF NOT EXISTS idx_sessions_updated_at ON sessions(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_log(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_session_id ON audit_log(session_id);
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_log(action);
CREATE INDEX IF NOT EXISTS idx_audit_surface ON audit_log(surface);
CREATE INDEX IF NOT EXISTS idx_audit_tool ON audit_log(tool);

-- Memory entries table for RAG system
CREATE TABLE IF NOT EXISTS memory_entries (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    date TEXT,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    access_count INTEGER DEFAULT 0,
    last_accessed DATETIME
);

-- Memory sources (which files/tools contributed to this memory)
CREATE TABLE IF NOT EXISTS memory_sources (
    id TEXT PRIMARY KEY,
    entry_id TEXT,
    source_type TEXT,
    source_path TEXT,
    FOREIGN KEY (entry_id) REFERENCES memory_entries(id) ON DELETE CASCADE
);

-- Vector embeddings (placeholder for future embedding provider plugin)
CREATE TABLE IF NOT EXISTS memory_vectors (
    entry_id TEXT PRIMARY KEY,
    embedding BLOB,
    FOREIGN KEY (entry_id) REFERENCES memory_entries(id) ON DELETE CASCADE
);

-- Indexes for memory tables
CREATE INDEX IF NOT EXISTS idx_memory_type ON memory_entries(type);
CREATE INDEX IF NOT EXISTS idx_memory_date ON memory_entries(date);
CREATE INDEX IF NOT EXISTS idx_memory_created_at ON memory_entries(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_memory_sources_entry_id ON memory_sources(entry_id);
`
