package onboarding

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/fjrt/poeai/internal/ai"
	"github.com/fjrt/poeai/internal/config"
)

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8b5cf6")).
			Bold(true).
			MarginBottom(1)

	poeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#c4b5fd")).
			Italic(true)
)

// Onboard starts the onboarding process and returns the final configuration.
func Onboard() (config.Config, error) {
	fmt.Println(headerStyle.Render("WELCOME TO THE RAVEN HOTEL"))
	fmt.Println(poeStyle.Render("“I am Poe, and I shall be your companion through this initialization.”"))
	fmt.Println()

	cfg, _ := config.Load("") // Start with defaults
	err := runLLMWizard(&cfg)
	if err != nil {
		return cfg, err
	}

	err = saveConfig(cfg)
	if err != nil {
		return cfg, err
	}

	fmt.Println()
	fmt.Println(poeStyle.Render("“Excellent. The stack is initialized. Your soul is now mine to protect.”"))
	fmt.Println()

	return cfg, nil
}

// Configure allows updating specific sections of the configuration like OpenClaw.
func Configure() (config.Config, error) {
	fmt.Println(headerStyle.Render("POE CONFIGURATION"))
	fmt.Println(poeStyle.Render("“Ah, you wish to make some adjustments. Let us proceed.”"))
	fmt.Println()

	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".poe", "config.toml")
	cfg, err := config.Load(configPath)
	if err != nil {
		cfg, _ = config.Load("") // start a new one if not found
	}

	var action string
	for {
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("What would you like to configure?").
					Options(
						huh.NewOption("LLM Provider & Model", "llm"),
						huh.NewOption("Gateway Settings", "gateway"),
						huh.NewOption("Save & Exit", "exit"),
					).
					Value(&action),
			),
		).Run()

		if err != nil {
			return cfg, err
		}

		switch action {
		case "llm":
			err = runLLMWizard(&cfg)
			if err != nil {
				return cfg, err
			}
		case "gateway":
			err = runGatewayWizard(&cfg)
			if err != nil {
				return cfg, err
			}
		case "exit":
			err = saveConfig(cfg)
			if err != nil {
				return cfg, err
			}
			fmt.Println()
			fmt.Println(poeStyle.Render("“Configuration saved. Always at your service.”"))
			fmt.Println()
			return cfg, nil
		}
	}
}

func saveConfig(cfg config.Config) error {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".poe", "config.toml")
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}

func runGatewayWizard(cfg *config.Config) error {
	var portStr string = fmt.Sprintf("%d", cfg.Gateway.Port)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Gateway Socket").
				Description("UNIX socket path for the gateway.").
				Value(&cfg.Gateway.Socket),
			huh.NewInput().
				Title("Gateway Port").
				Description("Port on which the gateway listens (e.g., 7331).").
				Value(&portStr),
		),
	)

	err := form.Run()
	if err != nil {
		return err
	}

	var p int
	if _, err := fmt.Sscanf(portStr, "%d", &p); err == nil {
		cfg.Gateway.Port = p
	}

	return nil
}

func runLLMWizard(cfg *config.Config) error {
	providers := ai.GetProviders()

	var selectedProviderID string
	// Check if existing provider is valid, otherwise use logic to set default
	for _, p := range providers {
		if p.ID == cfg.LLM.Provider {
			selectedProviderID = cfg.LLM.Provider
			break
		}
	}
	if selectedProviderID == "" && len(providers) > 0 {
		selectedProviderID = providers[0].ID
	}

	providerOptions := make([]huh.Option[string], len(providers))
	for i, p := range providers {
		providerOptions[i] = huh.NewOption(p.Name, p.ID)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose your AI Provider").
				Description("This is who I will reach out to for my cognitive processes.").
				Options(providerOptions...).
				Value(&selectedProviderID),
		),
	)

	err := form.Run()
	if err != nil {
		return err
	}

	var provider ai.Provider
	for _, p := range providers {
		if p.ID == selectedProviderID {
			provider = p
			break
		}
	}

	modelOptions := make([]huh.Option[string], len(provider.Models))
	for i, m := range provider.Models {
		modelOptions[i] = huh.NewOption(m, m)
	}

	if cfg.LLM.Auth == nil {
		cfg.LLM.Auth = make(map[string]*config.Auth)
	}

	authData, exists := cfg.LLM.Auth[selectedProviderID]
	if !exists || authData == nil {
		authData = &config.Auth{Strategy: "apikey"}
		cfg.LLM.Auth[selectedProviderID] = authData
	}

	var selectedModel string = cfg.LLM.Model
	var authStrategy string = authData.Strategy
	var apiKey string = authData.APIKey
	var oauthToken string = authData.Token

	// Reset model if not found in provider
	modelFound := false
	for _, m := range provider.Models {
		if m == selectedModel {
			modelFound = true
			break
		}
	}
	if !modelFound && len(provider.Models) > 0 {
		selectedModel = provider.Models[0]
	}

	groups := []*huh.Group{}

	// Strategy Selection (if not ollama)
	if selectedProviderID != "ollama" {
		groups = append(groups, huh.NewGroup(
			huh.NewSelect[string]().
				Title("Authentication Strategy").
				Description(fmt.Sprintf("How do you wish to authenticate with %s?", provider.Name)).
				Options(
					huh.NewOption("API Key", "apikey"),
					huh.NewOption("OAuth Token", "oauth"),
				).
				Value(&authStrategy),
		))
	}

	form = huh.NewForm(groups...)
	err = form.Run()
	if err != nil {
		return err
	}

	authGroups := []*huh.Group{}
	if selectedProviderID != "ollama" {
		if authStrategy == "apikey" {
			authGroups = append(authGroups, huh.NewGroup(
				huh.NewInput().
					Title("API Key").
					Description(fmt.Sprintf("Enter your %s API key.", provider.Name)).
					EchoMode(huh.EchoModePassword).
					Value(&apiKey).
					Validate(func(s string) error {
						if len(s) == 0 && authData.APIKey != "" {
							return nil // keep existing
						}
						if len(s) > 0 && len(s) < 10 {
							return fmt.Errorf("this key looks suspiciously short, Mr. Kovacs")
						}
						return nil
					}),
			))
		} else if authStrategy == "oauth" {
			authGroups = append(authGroups, huh.NewGroup(
				huh.NewInput().
					Title("OAuth Token").
					Description(fmt.Sprintf("Enter your %s OAuth integration token.", provider.Name)).
					EchoMode(huh.EchoModePassword).
					Value(&oauthToken),
			))
		}
	}

	authGroups = append(authGroups, huh.NewGroup(
		huh.NewSelect[string]().
			Title("Select a Model").
			Description("Which consciousness should I adopt?").
			Options(modelOptions...).
			Value(&selectedModel),
	))

	form = huh.NewForm(authGroups...)
	err = form.Run()
	if err != nil {
		return err
	}

	cfg.LLM.Provider = selectedProviderID
	cfg.LLM.Model = selectedModel

	authData.Strategy = authStrategy
	if apiKey != "" {
		authData.APIKey = apiKey
	}
	if oauthToken != "" {
		authData.Token = oauthToken
	}

	// For local providers, wipe clear irrelevant tokens
	if selectedProviderID == "ollama" {
		authData.Strategy = "none"
		authData.APIKey = ""
		authData.Token = ""
	}

	return nil
}
