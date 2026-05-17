// Example: Skills from Filesystem
//
// Loads Agent Skills (agentskills.io format) from a directory on disk.
// Demonstrates scanning for SKILL.md files, registering them, and
// activating a skill to augment the agent's behavior.
//
// Usage:
//
//	export OPENAI_API_KEY=sk-...
//	go run examples/skills/main.go
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cyberYiz/agent-sdk/pkg/agent"
	"github.com/cyberYiz/agent-sdk/pkg/llm/openai"
	"github.com/cyberYiz/agent-sdk/pkg/memory"
	"github.com/cyberYiz/agent-sdk/pkg/skills"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: OPENAI_API_KEY not set")
		os.Exit(1)
	}

	// Resolve the skills directory (relative to this example)
	skillsDir := filepath.Join("examples", "skills")
	absDir, err := filepath.Abs(skillsDir)
	if err != nil {
		absDir = skillsDir
	}

	fmt.Printf("Scanning for skills in: %s\n", absDir)

	// Scan and load all skills from the directory
	loaded, err := skills.LoadSkillsFromDirectory(absDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning skills: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d skill(s):\n", len(loaded))
	for name, s := range loaded {
		fmt.Printf("  - %s: %s\n", name, s.Description())
	}

	// Load a single skill directly
	pdfSkill, err := skills.LoadSkillFromPath(filepath.Join(absDir, "pdf-processing"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading pdf-processing: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("Loaded skill directly: %s\n", pdfSkill.Name())
	fmt.Printf("  Category: %s\n", pdfSkill.Category())
	fmt.Printf("  License: %s\n", pdfSkill.SkillFile().License)

	// Demonstrate reading a reference file
	refData, err := pdfSkill.ReadReference("FORMS.md")
	if err != nil {
		fmt.Printf("  Reference FORMS.md: not found (%v)\n", err)
	} else {
		fmt.Printf("  Reference FORMS.md: %d bytes\n", len(refData))
	}

	// Create agent and register the skill
	llm := openai.NewClient(apiKey, openai.WithModel("gpt-4o-mini"))
	mem := memory.NewConversationBuffer()

	sr := skills.NewRegistry()
	if err := sr.Register(pdfSkill); err != nil {
		panic(err)
	}
	if err := sr.Activate("pdf-processing"); err != nil {
		panic(err)
	}

	ag, err := agent.NewAgent(
		agent.WithLLM(llm),
		agent.WithMemory(mem),
		agent.WithSkillRegistry(sr),
		agent.WithSystemPrompt("You are a helpful assistant."),
		agent.WithName("SkillLoader"),
		agent.WithMaxIterations(5),
	)
	if err != nil {
		panic(err)
	}

	// Show how the skill augments the system prompt
	fmt.Println()
	fmt.Println("=== Combined System Prompt (first 300 chars) ===")
	full := ag.SystemPrompt()
	if len(full) > 300 {
		full = full[:300] + "..."
	}
	fmt.Println(full)
	fmt.Println("================================================")
	fmt.Println()

	// Run a PDF-related query
	ctx := context.Background()
	resp, err := ag.Run(ctx, "How do I extract text from a PDF file?")
	if err != nil {
		panic(err)
	}
	fmt.Println("=== Agent Response ===")
	fmt.Println(resp)
}
