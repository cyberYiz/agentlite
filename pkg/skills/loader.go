// Package skills provides the Agent Skills format loader (agentskills.io).
package skills

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/user/agent-sdk/pkg/interfaces"
)

// SkillFile represents a parsed SKILL.md file.
type SkillFile struct {
	// Frontmatter fields
	Name         string            `yaml:"name"`
	Description  string            `yaml:"description"`
	License      string            `yaml:"license"`
	Compatibility string           `yaml:"compatibility"`
	Metadata     map[string]string `yaml:"metadata"`
	AllowedTools string            `yaml:"allowed-tools"`

	// Derived fields
	Body      string // Markdown body after frontmatter
	DirPath   string // Absolute path to the skill directory
	FilePath  string // Absolute path to SKILL.md
}

// FileSkill wraps a SkillFile to implement interfaces.Skill.
type FileSkill struct {
	file      *SkillFile
	category  string
	tools     []interfaces.Tool
}

// NewFileSkill creates a FileSkill from a parsed SkillFile.
func NewFileSkill(file *SkillFile, opts ...FileSkillOption) *FileSkill {
	fs := &FileSkill{
		file:     file,
		category: "general",
	}
	for _, o := range opts {
		o(fs)
	}
	return fs
}

// FileSkillOption configures a FileSkill.
type FileSkillOption func(*FileSkill)

// WithSkillCategory sets the category.
func WithSkillCategory(cat string) FileSkillOption {
	return func(fs *FileSkill) { fs.category = cat }
}

// WithSkillTools attaches tools to the skill.
func WithSkillTools(tools ...interfaces.Tool) FileSkillOption {
	return func(fs *FileSkill) { fs.tools = tools }
}

// --- interfaces.Skill implementation ---

func (fs *FileSkill) Name() string               { return fs.file.Name }
func (fs *FileSkill) Description() string         { return fs.file.Description }
func (fs *FileSkill) Category() string            { return fs.category }

// SystemPromptAugment returns the full SKILL.md body as system prompt augmentation.
// This follows the agentskills progressive disclosure pattern:
// the body is loaded when the agent activates the skill.
func (fs *FileSkill) SystemPromptAugment() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Skill: %s\n", fs.file.Name))
	sb.WriteString(fs.file.Body)
	return sb.String()
}

func (fs *FileSkill) Tools() []interfaces.Tool { return fs.tools }

// Init validates the skill directory exists.
func (fs *FileSkill) Init(_ context.Context) error {
	if fs.file.DirPath == "" {
		return nil // in-memory skill, no directory to validate
	}
	info, err := os.Stat(fs.file.DirPath)
	if err != nil {
		return fmt.Errorf("skill directory %q: %w", fs.file.DirPath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("skill path %q is not a directory", fs.file.DirPath)
	}
	return nil
}

func (fs *FileSkill) PreRun(_ context.Context, prompt string) (string, error) {
	return prompt, nil
}
func (fs *FileSkill) PostRun(_ context.Context, response string) (string, error) {
	return response, nil
}

// --- Additional FileSkill methods ---

// SkillFile returns the underlying parsed SKILL.md data.
func (fs *FileSkill) SkillFile() *SkillFile { return fs.file }

// ReadReference reads a file from the skill's references/ directory.
func (fs *FileSkill) ReadReference(name string) ([]byte, error) {
	return fs.readFile("references", name)
}

// ReadScript reads a file from the skill's scripts/ directory.
func (fs *FileSkill) ReadScript(name string) ([]byte, error) {
	return fs.readFile("scripts", name)
}

// ReadAsset reads a file from the skill's assets/ directory.
func (fs *FileSkill) ReadAsset(name string) ([]byte, error) {
	return fs.readFile("assets", name)
}

func (fs *FileSkill) readFile(subdir, name string) ([]byte, error) {
	// Prevent path traversal
	clean := filepath.Clean(name)
	if strings.Contains(clean, "..") {
		return nil, fmt.Errorf("invalid path: %s", name)
	}
	path := filepath.Join(fs.file.DirPath, subdir, clean)
	return os.ReadFile(path)
}

// --- Loader functions ---

// LoadSkillFromPath loads a single skill from a directory containing SKILL.md.
func LoadSkillFromPath(dirPath string, opts ...FileSkillOption) (*FileSkill, error) {
	file, err := parseSkillFile(dirPath)
	if err != nil {
		return nil, err
	}
	return NewFileSkill(file, opts...), nil
}

// LoadSkillsFromDirectory scans a directory recursively and loads all
// subdirectories that contain a SKILL.md file.
// Returns skills keyed by their name.
func LoadSkillsFromDirectory(root string, opts ...FileSkillOption) (map[string]*FileSkill, error) {
	skills := make(map[string]*FileSkill)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		// Check if this directory has SKILL.md
		skillPath := filepath.Join(path, "SKILL.md")
		if _, statErr := os.Stat(skillPath); statErr == nil {
			file, parseErr := parseSkillFile(path)
			if parseErr != nil {
				// Skip invalid skills but don't fail the whole scan
				return nil
			}
			// Validate name matches directory
			dirName := filepath.Base(path)
			if file.Name != dirName {
				return nil // silently skip mismatched names per spec
			}
			skills[file.Name] = NewFileSkill(file, opts...)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan skills directory %q: %w", root, err)
	}
	return skills, nil
}

// parseSkillFile reads and parses a SKILL.md file from the given directory.
func parseSkillFile(dirPath string) (*SkillFile, error) {
	skillPath := filepath.Join(dirPath, "SKILL.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		return nil, fmt.Errorf("read SKILL.md in %q: %w", dirPath, err)
	}

	frontmatter, body, err := splitFrontmatter(data)
	if err != nil {
		return nil, fmt.Errorf("parse frontmatter in %q: %w", dirPath, err)
	}

	var sf SkillFile
	if err := yaml.Unmarshal(frontmatter, &sf); err != nil {
		return nil, fmt.Errorf("yaml parse in %q: %w", dirPath, err)
	}

	// Validate required fields
	if sf.Name == "" {
		return nil, fmt.Errorf("SKILL.md in %q: missing required field 'name'", dirPath)
	}
	if sf.Description == "" {
		return nil, fmt.Errorf("SKILL.md in %q: missing required field 'description'", dirPath)
	}

	// Validate name format per spec: lowercase, numbers, hyphens only
	if err := validateSkillName(sf.Name); err != nil {
		return nil, fmt.Errorf("SKILL.md in %q: %w", dirPath, err)
	}

	sf.Body = string(bytes.TrimSpace(body))
	sf.DirPath = dirPath
	sf.FilePath = skillPath

	return &sf, nil
}

// splitFrontmatter splits YAML frontmatter (between --- delimiters) from body.
func splitFrontmatter(data []byte) ([]byte, []byte, error) {
	text := string(data)
	// Must start with "---\n" or "---\r\n"
	if !strings.HasPrefix(text, "---") {
		return nil, nil, fmt.Errorf("missing YAML frontmatter (must start with ---)")
	}

	// Find closing "---"
	// Skip the opening ---
	rest := text[3:]
	// Handle \n after opening ---
	if strings.HasPrefix(rest, "\n") {
		rest = rest[1:]
	} else if strings.HasPrefix(rest, "\r\n") {
		rest = rest[2:]
	}

	// Find closing --- on its own line
	endIdx := strings.Index(rest, "\n---")
	if endIdx < 0 {
		return nil, nil, fmt.Errorf("missing closing --- for YAML frontmatter")
	}

	frontmatter := []byte(rest[:endIdx])
	body := []byte(rest[endIdx+4:]) // skip "\n---"
	return frontmatter, body, nil
}

// validateSkillName checks the name follows the agentskills spec:
// 1-64 chars, lowercase letters/numbers/hyphens, no leading/trailing hyphens, no consecutive hyphens.
func validateSkillName(name string) error {
	if len(name) < 1 || len(name) > 64 {
		return fmt.Errorf("name must be 1-64 characters, got %d", len(name))
	}
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return fmt.Errorf("name must not start or end with hyphen")
	}
	if strings.Contains(name, "--") {
		return fmt.Errorf("name must not contain consecutive hyphens")
	}
	for _, r := range name {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '-' {
			continue
		}
		return fmt.Errorf("name contains invalid character %q (only lowercase letters, numbers, hyphens allowed)", r)
	}
	return nil
}

// --- Prompt integration: generate <available_skills> XML ---

// ToPromptXML generates the recommended <available_skills> XML block
// for agent system prompts, following the agentskills format.
func ToPromptXML(skills []*FileSkill) string {
	var sb strings.Builder
	sb.WriteString("<available_skills>\n")
	for _, s := range skills {
		loc := s.file.FilePath
		if loc == "" {
			loc = fmt.Sprintf("./%s/SKILL.md", s.file.Name)
		}
		sb.WriteString("<skill>\n")
		sb.WriteString(fmt.Sprintf("<name>\n%s\n</name>\n", s.file.Name))
		sb.WriteString(fmt.Sprintf("<description>\n%s\n</description>\n", s.file.Description))
		sb.WriteString(fmt.Sprintf("<location>\n%s\n</location>\n", loc))
		sb.WriteString("</skill>\n")
	}
	sb.WriteString("</available_skills>")
	return sb.String()
}

// compile-time check
var _ interfaces.Skill = (*FileSkill)(nil)
