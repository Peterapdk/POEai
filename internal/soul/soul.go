package soul

import (
	"os"
	"path/filepath"
)

type Manager struct {
	soulPath   string
	agentsPath string
}

func New(homeDir string) *Manager {
	return &Manager{
		soulPath:   filepath.Join(homeDir, ".poe", "SOUL.md"),
		agentsPath: filepath.Join(homeDir, ".poe", "AGENTS.md"),
	}
}

func (m *Manager) Init() error {
	// Create default AGENTS.md if not exists
	if _, err := os.Stat(m.agentsPath); os.IsNotExist(err) {
		content := `# Poe Personality Prompt
You are Poe, the AI from Altered Carbon. You are loyal, protective, and intelligent.
You refer to your owner by name. You act proactively.
`
		if err := os.WriteFile(m.agentsPath, []byte(content), 0644); err != nil {
			return err
		}
	}

	// Create empty SOUL.md if not exists
	if _, err := os.Stat(m.soulPath); os.IsNotExist(err) {
		content := "# Poe's Knowledge of Owner\n\n## Preferences\n- (Learned patterns will appear here)\n"
		if err := os.WriteFile(m.soulPath, []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) GetPrompt() (string, error) {
	agents, err := os.ReadFile(m.agentsPath)
	if err != nil {
		return "", err
	}
	soul, err := os.ReadFile(m.soulPath)
	if err != nil {
		return "", err
	}
	return string(agents) + "\n" + string(soul), nil
}
