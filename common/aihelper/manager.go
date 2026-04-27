package aihelper

import (
	"context"
	"sync"
)

var ctx = context.Background()

// AIHelperManager stores user/session scoped helpers.
type AIHelperManager struct {
	helpers map[string]map[string]*AIHelper
	mu      sync.RWMutex
}

// NewAIHelperManager creates a manager instance.
func NewAIHelperManager() *AIHelperManager {
	return &AIHelperManager{
		helpers: make(map[string]map[string]*AIHelper),
	}
}

// GetOrCreateAIHelper returns the session helper and switches model type when requested.
func (m *AIHelperManager) GetOrCreateAIHelper(userName string, sessionID string, modelType string, config map[string]interface{}) (*AIHelper, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	userHelpers, exists := m.helpers[userName]
	if !exists {
		userHelpers = make(map[string]*AIHelper)
		m.helpers[userName] = userHelpers
	}

	factory := GetGlobalFactory()

	helper, exists := userHelpers[sessionID]
	if exists {
		if helper.GetModelType() == modelType {
			return helper, nil
		}
		newModel, err := factory.CreateAIModel(ctx, modelType, config)
		if err != nil {
			return nil, err
		}
		helper.SetModel(newModel)
		return helper, nil
	}

	helper, err := factory.CreateAIHelper(ctx, modelType, sessionID, config)
	if err != nil {
		return nil, err
	}

	userHelpers[sessionID] = helper
	return helper, nil
}

// GetAIHelper returns an existing session helper.
func (m *AIHelperManager) GetAIHelper(userName string, sessionID string) (*AIHelper, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userHelpers, exists := m.helpers[userName]
	if !exists {
		return nil, false
	}

	helper, exists := userHelpers[sessionID]
	return helper, exists
}

// RemoveAIHelper removes a session helper.
func (m *AIHelperManager) RemoveAIHelper(userName string, sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	userHelpers, exists := m.helpers[userName]
	if !exists {
		return
	}

	delete(userHelpers, sessionID)

	if len(userHelpers) == 0 {
		delete(m.helpers, userName)
	}
}

// GetUserSessions returns cached session IDs for a user.
func (m *AIHelperManager) GetUserSessions(userName string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userHelpers, exists := m.helpers[userName]
	if !exists {
		return []string{}
	}

	sessionIDs := make([]string, 0, len(userHelpers))
	for sessionID := range userHelpers {
		sessionIDs = append(sessionIDs, sessionID)
	}

	return sessionIDs
}

var globalManager *AIHelperManager
var once sync.Once

// GetGlobalManager returns the global manager.
func GetGlobalManager() *AIHelperManager {
	once.Do(func() {
		globalManager = NewAIHelperManager()
	})
	return globalManager
}
