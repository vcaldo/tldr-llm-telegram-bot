package llm

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Prompts struct {
	Summary     map[string]string `toml:"summary"`
	Problematic map[string]string `toml:"problematic"`
}

var prompts Prompts

func LoadPrompts(filePath string) error {
	if _, err := toml.DecodeFile(filePath, &prompts); err != nil {
		return fmt.Errorf("failed to load prompts from TOML file: %v", err)
	}
	return nil
}

func GetPrompt(category, lang string) (string, error) {
	switch category {
	case "summary":
		if prompt, ok := prompts.Summary[lang]; ok {
			return prompt, nil
		}
	case "problematic":
		if prompt, ok := prompts.Problematic[lang]; ok {
			return prompt, nil
		}
	}
	return "", fmt.Errorf("prompt not found for category '%s' and language '%s'", category, lang)
}
