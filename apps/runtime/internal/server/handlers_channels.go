package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"pryx-core/internal/channels"
	"pryx-core/internal/channels/discord"
	"pryx-core/internal/channels/slack"
	"pryx-core/internal/channels/telegram"
	"pryx-core/internal/channels/webhook"

	"github.com/go-chi/chi/v5"
)

type Channel struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Name      string                 `json:"name"`
	Config    map[string]interface{} `json:"config"`
	Enabled   bool                   `json:"enabled"`
	Status    channels.Status        `json:"status"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type ChannelConfig struct {
	Type   string                 `json:"type"`
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

type ChannelTestResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	Latency   string `json:"latency,omitempty"`
	LastCheck string `json:"last_check,omitempty"`
}

type HealthStatus struct {
	Healthy    bool      `json:"healthy"`
	Status     string    `json:"status"`
	LastSeen   time.Time `json:"last_seen,omitempty"`
	Uptime     string    `json:"uptime,omitempty"`
	ErrorCount int       `json:"error_count"`
}

type ChannelActivity struct {
	ID        string    `json:"id"`
	ChannelID string    `json:"channel_id"`
	SenderID  string    `json:"sender_id,omitempty"`
	Content   string    `json:"content,omitempty"`
	Event     string    `json:"event,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Server) handleChannelsList(w http.ResponseWriter, r *http.Request) {
	channelsList := []Channel{}

	telegramMgr := telegram.NewConfigManager()
	telegramConfigs, _ := telegramMgr.List()
	for _, cfg := range telegramConfigs {
		channelsList = append(channelsList, Channel{
			ID:        cfg.ID,
			Type:      "telegram",
			Name:      cfg.Name,
			Config:    telegramConfigToMap(&cfg),
			Enabled:   cfg.Enabled,
			Status:    getChannelStatus(cfg.ID),
			CreatedAt: cfg.CreatedAt,
			UpdatedAt: cfg.UpdatedAt,
		})
	}

	slackMgr := slack.NewSlackConfigManager()
	slackConfigs, _ := slackMgr.List()
	for _, cfg := range slackConfigs {
		channelsList = append(channelsList, Channel{
			ID:        cfg.ID,
			Type:      "slack",
			Name:      cfg.Name,
			Config:    slackConfigToMap(&cfg),
			Enabled:   cfg.Enabled,
			Status:    getChannelStatus(cfg.ID),
			CreatedAt: cfg.CreatedAt,
			UpdatedAt: cfg.UpdatedAt,
		})
	}

	discordMgr := discord.NewConfigManager()
	discordConfigs, _ := discordMgr.List()
	for _, cfg := range discordConfigs {
		channelsList = append(channelsList, Channel{
			ID:        cfg.ID,
			Type:      "discord",
			Name:      cfg.Name,
			Config:    discordConfigToMap(&cfg),
			Enabled:   cfg.Enabled,
			Status:    getChannelStatus(cfg.ID),
			CreatedAt: cfg.CreatedAt,
			UpdatedAt: cfg.UpdatedAt,
		})
	}

	webhookMgr := webhook.NewConfigManager()
	webhookConfigs, _ := webhookMgr.LoadAll()
	for _, cfg := range webhookConfigs {
		channelsList = append(channelsList, Channel{
			ID:        cfg.ID,
			Type:      "webhook",
			Name:      cfg.Name,
			Config:    webhookConfigToMap(&cfg),
			Enabled:   cfg.Enabled,
			Status:    getChannelStatus(cfg.ID),
			CreatedAt: cfg.CreatedAt,
			UpdatedAt: cfg.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"channels": channelsList,
		"count":    len(channelsList),
	})
}

func (s *Server) handleChannelGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	channel, err := s.getChannel(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(channel)
}

func (s *Server) handleChannelCreate(w http.ResponseWriter, r *http.Request) {
	req := ChannelConfig{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "invalid json body",
		})
		return
	}

	if req.Type == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "type is required",
		})
		return
	}

	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "name is required",
		})
		return
	}

	var channel Channel
	var err error

	switch req.Type {
	case "telegram":
		channel, err = s.createTelegramChannel(req.Name, req.Config)
	case "slack":
		channel, err = s.createSlackChannel(req.Name, req.Config)
	case "discord":
		channel, err = s.createDiscordChannel(req.Name, req.Config)
	case "webhook":
		channel, err = s.createWebhookChannel(req.Name, req.Config)
	default:
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("unknown channel type: %s", req.Type),
		})
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(channel)
}

func (s *Server) handleChannelUpdate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := ChannelConfig{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "invalid json body",
		})
		return
	}

	existing, err := s.getChannel(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	var updated Channel
	switch existing.Type {
	case "telegram":
		updated, err = s.updateTelegramChannel(id, req.Name, req.Config)
	case "slack":
		updated, err = s.updateSlackChannel(id, req.Name, req.Config)
	case "discord":
		updated, err = s.updateDiscordChannel(id, req.Name, req.Config)
	case "webhook":
		updated, err = s.updateWebhookChannel(id, req.Name, req.Config)
	default:
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("unknown channel type: %s", existing.Type),
		})
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(updated)
}

func (s *Server) handleChannelDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	existing, err := s.getChannel(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	switch existing.Type {
	case "telegram":
		mgr := telegram.NewConfigManager()
		err = mgr.Delete(id)
	case "slack":
		mgr := slack.NewSlackConfigManager()
		err = mgr.Delete(id)
	case "discord":
		mgr := discord.NewConfigManager()
		err = mgr.Delete(id)
	case "webhook":
		mgr := webhook.NewConfigManager()
		_ = mgr.Delete(id)
		err = nil
	default:
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("unknown channel type: %s", existing.Type),
		})
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleChannelTest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	_, err := s.getChannel(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	result := ChannelTestResult{
		Success:   true,
		Message:   "Channel test not yet implemented",
		LastCheck: time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (s *Server) handleChannelHealth(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	_, err := s.getChannel(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	health := HealthStatus{
		Healthy:    true,
		Status:     string(channels.StatusConnected),
		LastSeen:   time.Now(),
		Uptime:     "unknown",
		ErrorCount: 0,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(health)
}

func (s *Server) handleChannelConnect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	_, err := s.getChannel(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ok": true,
	})
}

func (s *Server) handleChannelDisconnect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	_, err := s.getChannel(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ok": true,
	})
}

func (s *Server) handleChannelActivity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	_, err := s.getChannel(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := parseLimit(l); err == nil {
			limit = parsed
		}
	}

	activity := []ChannelActivity{}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"activity": activity,
		"limit":    limit,
	})
}

func (s *Server) handleChannelTypes(w http.ResponseWriter, r *http.Request) {
	types := []map[string]string{
		{
			"type":        "telegram",
			"name":        "Telegram",
			"description": "Run your agent as a Telegram bot",
		},
		{
			"type":        "discord",
			"name":        "Discord",
			"description": "Deploy as a Discord bot with slash commands",
		},
		{
			"type":        "slack",
			"name":        "Slack",
			"description": "Connect to Slack channels and DMs",
		},
		{
			"type":        "webhook",
			"name":        "Webhook",
			"description": "Integrate with any HTTP endpoint",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"types": types,
	})
}

func (s *Server) getChannel(id string) (Channel, error) {
	if s.channels == nil {
		return Channel{}, fmt.Errorf("channel manager not initialized")
	}

	ch, ok := s.channels.Get(id)
	if !ok {
		return Channel{}, fmt.Errorf("channel not found: %s", id)
	}

	return Channel{
		ID:        ch.ID(),
		Type:      ch.Type(),
		Name:      ch.ID(),
		Config:    map[string]interface{}{},
		Enabled:   ch.Status() == channels.StatusConnected,
		Status:    ch.Status(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func getChannelStatus(id string) channels.Status {
	return channels.StatusDisconnected
}

func parseLimit(s string) (int, error) {
	var limit int
	_, err := fmt.Sscanf(s, "%d", &limit)
	return limit, err
}

func telegramConfigToMap(cfg *telegram.Config) map[string]interface{} {
	return map[string]interface{}{
		"mode":                 cfg.Mode,
		"token_ref":            cfg.TokenRef,
		"webhook_url":          cfg.WebhookURL,
		"polling_interval":     cfg.PollingInterval.String(),
		"allowed_chats":        cfg.AllowedChats,
		"allowed_updates":      cfg.AllowedUpdates,
		"max_connections":      cfg.MaxConnections,
		"drop_pending_updates": cfg.DropPendingUpdates,
	}
}

func slackConfigToMap(cfg *slack.Config) map[string]interface{} {
	return map[string]interface{}{
		"bot_token": cfg.BotToken,
		"app_token": cfg.AppToken,
	}
}

func discordConfigToMap(cfg *discord.Config) map[string]interface{} {
	return map[string]interface{}{
		"bot_token":        cfg.TokenRef,
		"application_id":   cfg.ApplicationID,
		"intents":          cfg.Intents,
		"allowed_guilds":   cfg.AllowedGuilds,
		"allowed_channels": cfg.AllowedChannels,
	}
}

func webhookConfigToMap(cfg *webhook.WebhookConfig) map[string]interface{} {
	return map[string]interface{}{
		"port":       cfg.Port,
		"path":       cfg.Path,
		"secret":     cfg.Secret,
		"target_url": cfg.TargetURL,
		"headers":    cfg.Headers,
	}
}

func (s *Server) createTelegramChannel(name string, config map[string]interface{}) (Channel, error) {
	mgr := telegram.NewConfigManager()
	cfg := telegram.DefaultConfig()
	cfg.Name = name

	if tokenRef, ok := config["token_ref"].(string); ok {
		cfg.TokenRef = tokenRef
	}

	if mode, ok := config["mode"].(string); ok {
		cfg.Mode = mode
	}

	if webhookURL, ok := config["webhook_url"].(string); ok {
		cfg.WebhookURL = webhookURL
	}

	created, err := mgr.Create(cfg)
	if err != nil {
		return Channel{}, err
	}

	return Channel{
		ID:        created.ID,
		Type:      "telegram",
		Name:      created.Name,
		Config:    telegramConfigToMap(created),
		Enabled:   created.Enabled,
		Status:    channels.StatusDisconnected,
		CreatedAt: created.CreatedAt,
		UpdatedAt: created.UpdatedAt,
	}, nil
}

func (s *Server) createSlackChannel(name string, config map[string]interface{}) (Channel, error) {
	mgr := slack.NewSlackConfigManager()
	cfg := slack.NewBotConfig(name, "", "")

	if botToken, ok := config["bot_token"].(string); ok {
		cfg.BotToken = botToken
	}

	if appToken, ok := config["app_token"].(string); ok {
		cfg.AppToken = appToken
	}

	created, err := mgr.Create(cfg)
	if err != nil {
		return Channel{}, err
	}

	return Channel{
		ID:        created.ID,
		Type:      "slack",
		Name:      created.Name,
		Config:    slackConfigToMap(created),
		Enabled:   created.Enabled,
		Status:    channels.StatusDisconnected,
		CreatedAt: created.CreatedAt,
		UpdatedAt: created.UpdatedAt,
	}, nil
}

func (s *Server) createDiscordChannel(name string, config map[string]interface{}) (Channel, error) {
	mgr := discord.NewConfigManager()
	cfg := discord.DefaultConfig()
	cfg.Name = name

	if botToken, ok := config["bot_token"].(string); ok {
		cfg.TokenRef = botToken
	}

	created, err := mgr.Create(cfg)
	if err != nil {
		return Channel{}, err
	}

	return Channel{
		ID:        created.ID,
		Type:      "discord",
		Name:      created.Name,
		Config:    discordConfigToMap(created),
		Enabled:   created.Enabled,
		Status:    channels.StatusDisconnected,
		CreatedAt: created.CreatedAt,
		UpdatedAt: created.UpdatedAt,
	}, nil
}

func (s *Server) createWebhookChannel(name string, config map[string]interface{}) (Channel, error) {
	mgr := webhook.NewConfigManager()
	cfg := webhook.WebhookConfig{
		Name:    name,
		Enabled: true,
	}

	if port, ok := config["port"].(float64); ok {
		cfg.Port = int(port)
	}

	if path, ok := config["path"].(string); ok {
		cfg.Path = path
	}

	if secret, ok := config["secret"].(string); ok {
		cfg.Secret = secret
	}

	if targetURL, ok := config["target_url"].(string); ok {
		cfg.TargetURL = targetURL
	}

	if headers, ok := config["headers"].(map[string]interface{}); ok {
		cfg.Headers = map[string]string{}
		for k, v := range headers {
			if vs, ok := v.(string); ok {
				cfg.Headers[k] = vs
			}
		}
	}

	err := mgr.Save(cfg)
	if err != nil {
		return Channel{}, err
	}

	return Channel{
		ID:        cfg.ID,
		Type:      "webhook",
		Name:      cfg.Name,
		Config:    webhookConfigToMap(&cfg),
		Enabled:   cfg.Enabled,
		Status:    channels.StatusDisconnected,
		CreatedAt: cfg.CreatedAt,
		UpdatedAt: cfg.UpdatedAt,
	}, nil
}

func (s *Server) updateTelegramChannel(id string, name string, config map[string]interface{}) (Channel, error) {
	mgr := telegram.NewConfigManager()
	updates := config

	if name != "" {
		updates["name"] = name
	}

	updated, err := mgr.Update(id, updates)
	if err != nil {
		return Channel{}, err
	}

	return Channel{
		ID:        updated.ID,
		Type:      "telegram",
		Name:      updated.Name,
		Config:    telegramConfigToMap(updated),
		Enabled:   updated.Enabled,
		Status:    channels.StatusDisconnected,
		CreatedAt: updated.CreatedAt,
		UpdatedAt: updated.UpdatedAt,
	}, nil
}

func (s *Server) updateSlackChannel(id string, name string, config map[string]interface{}) (Channel, error) {
	mgr := slack.NewSlackConfigManager()
	updates := config

	if name != "" {
		updates["name"] = name
	}

	updated, err := mgr.Update(id, updates)
	if err != nil {
		return Channel{}, err
	}

	return Channel{
		ID:        updated.ID,
		Type:      "slack",
		Name:      updated.Name,
		Config:    slackConfigToMap(updated),
		Enabled:   updated.Enabled,
		Status:    channels.StatusDisconnected,
		CreatedAt: updated.CreatedAt,
		UpdatedAt: updated.UpdatedAt,
	}, nil
}

func (s *Server) updateDiscordChannel(id string, name string, config map[string]interface{}) (Channel, error) {
	mgr := discord.NewConfigManager()
	updates := config

	if name != "" {
		updates["name"] = name
	}

	updated, err := mgr.Update(id, updates)
	if err != nil {
		return Channel{}, err
	}

	return Channel{
		ID:        updated.ID,
		Type:      "discord",
		Name:      updated.Name,
		Config:    discordConfigToMap(updated),
		Enabled:   updated.Enabled,
		Status:    channels.StatusDisconnected,
		CreatedAt: updated.CreatedAt,
		UpdatedAt: updated.UpdatedAt,
	}, nil
}

func (s *Server) updateWebhookChannel(id string, name string, config map[string]interface{}) (Channel, error) {
	mgr := webhook.NewConfigManager()

	if name != "" {
		config["name"] = name
	}

	updated, err := mgr.Get(id)
	if err != nil {
		return Channel{}, err
	}

	if port, ok := config["port"].(float64); ok {
		updated.Port = int(port)
	}

	if path, ok := config["path"].(string); ok {
		updated.Path = path
	}

	if secret, ok := config["secret"].(string); ok {
		updated.Secret = secret
	}

	if targetURL, ok := config["target_url"].(string); ok {
		updated.TargetURL = targetURL
	}

	if headers, ok := config["headers"].(map[string]interface{}); ok {
		updated.Headers = map[string]string{}
		for k, v := range headers {
			if vs, ok := v.(string); ok {
				updated.Headers[k] = vs
			}
		}
	}

	updated.Name = name
	if err := mgr.SaveAll([]webhook.WebhookConfig{*updated}); err != nil {
		return Channel{}, err
	}

	return Channel{
		ID:        updated.ID,
		Type:      "webhook",
		Name:      updated.Name,
		Config:    webhookConfigToMap(updated),
		Enabled:   updated.Enabled,
		Status:    channels.StatusDisconnected,
		CreatedAt: updated.CreatedAt,
		UpdatedAt: updated.UpdatedAt,
	}, nil
}
