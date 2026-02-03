package store

import (
	"database/sql"
	"time"
)

// MeshPairingSession represents a mesh pairing session
type MeshPairingSession struct {
	ID         string    `json:"id"`
	Code       string    `json:"code"`
	DeviceID   string    `json:"device_id"`
	DeviceName string    `json:"device_name"`
	ServerURL  string    `json:"server_url"`
	Nonce      string    `json:"nonce"`
	Status     string    `json:"status"` // pending, approved, rejected, expired
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// MeshDevice represents a paired mesh device
type MeshDevice struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	PublicKey string    `json:"public_key"`
	PairedAt  time.Time `json:"paired_at"`
	LastSeen  time.Time `json:"last_seen"`
	IsActive  bool      `json:"is_active"`
	Metadata  string    `json:"metadata"`
}

// MeshSyncEvent represents a mesh sync event
type MeshSyncEvent struct {
	ID             string    `json:"id"`
	EventType      string    `json:"event_type"`
	SourceDeviceID string    `json:"source_device_id"`
	TargetDeviceID string    `json:"target_device_id"`
	Payload        string    `json:"payload"`
	CreatedAt      time.Time `json:"created_at"`
}

// CreatePairingSession creates a new mesh pairing session
func (s *Store) CreatePairingSession(session *MeshPairingSession) error {
	_, err := s.DB.Exec(`
		INSERT INTO mesh_pairing_sessions (id, code, device_id, device_name, server_url, nonce, status, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		session.ID,
		session.Code,
		session.DeviceID,
		session.DeviceName,
		session.ServerURL,
		session.Nonce,
		session.Status,
		session.ExpiresAt,
		session.CreatedAt,
	)
	return err
}

// GetPairingSessionByCode retrieves a pairing session by its code
func (s *Store) GetPairingSessionByCode(code string) (*MeshPairingSession, error) {
	var session MeshPairingSession
	err := s.DB.QueryRow(`
		SELECT id, code, device_id, device_name, server_url, nonce, status, expires_at, created_at
		FROM mesh_pairing_sessions
		WHERE code = ?
	`, code).Scan(
		&session.ID,
		&session.Code,
		&session.DeviceID,
		&session.DeviceName,
		&session.ServerURL,
		&session.Nonce,
		&session.Status,
		&session.ExpiresAt,
		&session.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdatePairingSessionStatus updates the status of a pairing session
func (s *Store) UpdatePairingSessionStatus(code, status string) error {
	_, err := s.DB.Exec(`
		UPDATE mesh_pairing_sessions SET status = ? WHERE code = ?
	`, status, code)
	return err
}

// DeleteExpiredPairingSessions removes expired pairing sessions
func (s *Store) DeleteExpiredPairingSessions() error {
	_, err := s.DB.Exec(`DELETE FROM mesh_pairing_sessions WHERE expires_at < ?`, time.Now())
	return err
}

// CreateMeshDevice creates a new paired mesh device
func (s *Store) CreateMeshDevice(device *MeshDevice) error {
	_, err := s.DB.Exec(`
		INSERT INTO mesh_devices (id, name, public_key, paired_at, last_seen, is_active, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		device.ID,
		device.Name,
		device.PublicKey,
		device.PairedAt,
		device.LastSeen,
		device.IsActive,
		device.Metadata,
	)
	return err
}

// GetMeshDeviceByID retrieves a mesh device by ID
func (s *Store) GetMeshDeviceByID(id string) (*MeshDevice, error) {
	var device MeshDevice
	err := s.DB.QueryRow(`
		SELECT id, name, public_key, paired_at, last_seen, is_active, metadata
		FROM mesh_devices
		WHERE id = ?
	`, id).Scan(
		&device.ID,
		&device.Name,
		&device.PublicKey,
		&device.PairedAt,
		&device.LastSeen,
		&device.IsActive,
		&device.Metadata,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// ListMeshDevices lists all active mesh devices
func (s *Store) ListMeshDevices() ([]*MeshDevice, error) {
	rows, err := s.DB.Query(`
		SELECT id, name, public_key, paired_at, last_seen, is_active, metadata
		FROM mesh_devices
		WHERE is_active = 1
		ORDER BY paired_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*MeshDevice
	for rows.Next() {
		var device MeshDevice
		if err := rows.Scan(
			&device.ID,
			&device.Name,
			&device.PublicKey,
			&device.PairedAt,
			&device.LastSeen,
			&device.IsActive,
			&device.Metadata,
		); err != nil {
			return nil, err
		}
		devices = append(devices, &device)
	}
	return devices, rows.Err()
}

// UpdateMeshDeviceLastSeen updates the last seen timestamp for a device
func (s *Store) UpdateMeshDeviceLastSeen(id string) error {
	_, err := s.DB.Exec(`
		UPDATE mesh_devices SET last_seen = ? WHERE id = ?
	`, time.Now(), id)
	return err
}

// DeactivateMeshDevice marks a mesh device as inactive
func (s *Store) DeactivateMeshDevice(id string) error {
	_, err := s.DB.Exec(`
		UPDATE mesh_devices SET is_active = 0 WHERE id = ?
	`, id)
	return err
}

// CreateMeshSyncEvent creates a new mesh sync event
func (s *Store) CreateMeshSyncEvent(event *MeshSyncEvent) error {
	_, err := s.DB.Exec(`
		INSERT INTO mesh_sync_events (id, event_type, source_device_id, target_device_id, payload, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		event.ID,
		event.EventType,
		event.SourceDeviceID,
		event.TargetDeviceID,
		event.Payload,
		event.CreatedAt,
	)
	return err
}

// ListMeshSyncEvents lists mesh sync events with optional limits
func (s *Store) ListMeshSyncEvents(limit int) ([]*MeshSyncEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.DB.Query(`
		SELECT id, event_type, source_device_id, target_device_id, payload, created_at
		FROM mesh_sync_events
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*MeshSyncEvent
	for rows.Next() {
		var event MeshSyncEvent
		if err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.SourceDeviceID,
			&event.TargetDeviceID,
			&event.Payload,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, &event)
	}
	return events, rows.Err()
}
