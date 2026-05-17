// Package builtin provides built-in skills.
package builtin

import (
	"context"

	"github.com/cyberYiz/agentlite/pkg/interfaces"
)

// BaseSkill provides a convenient base for implementing skills.
// Embed it and override the methods you need.
type BaseSkill struct {
	name        string
	description string
	category    string
	promptAug   string
	tools       []interfaces.Tool
}

// NewBaseSkill creates a new base skill.
func NewBaseSkill(name, description, category string) *BaseSkill {
	return &BaseSkill{
		name:        name,
		description: description,
		category:    category,
	}
}

// WithPrompt sets the system prompt augmentation.
func (b *BaseSkill) WithPrompt(prompt string) *BaseSkill {
	b.promptAug = prompt
	return b
}

// WithTools sets the skill's tools.
func (b *BaseSkill) WithTools(tools ...interfaces.Tool) *BaseSkill {
	b.tools = tools
	return b
}

func (b *BaseSkill) Name() string                                        { return b.name }
func (b *BaseSkill) Description() string                                 { return b.description }
func (b *BaseSkill) Category() string                                    { return b.category }
func (b *BaseSkill) SystemPromptAugment() string                         { return b.promptAug }
func (b *BaseSkill) Tools() []interfaces.Tool                            { return b.tools }
func (b *BaseSkill) Init(_ context.Context) error                        { return nil }
func (b *BaseSkill) PreRun(_ context.Context, p string) (string, error)  { return p, nil }
func (b *BaseSkill) PostRun(_ context.Context, r string) (string, error) { return r, nil }

// --- Built-in skill factories ---

// NewCodeReviewSkill returns a skill for code review.
func NewCodeReviewSkill() interfaces.Skill {
	return NewBaseSkill(
		"code_review",
		"Expert code review and analysis",
		"code",
	).WithPrompt(`You are an expert code reviewer. When analyzing code:
1. Identify bugs, edge cases, and potential issues
2. Suggest performance improvements
3. Check for security vulnerabilities
4. Recommend best practices and design patterns
5. Provide clear, actionable feedback with examples`)
}

// NewDataAnalysisSkill returns a skill for data analysis.
func NewDataAnalysisSkill() interfaces.Skill {
	return NewBaseSkill(
		"data_analysis",
		"Data analysis and visualization",
		"data",
	).WithPrompt(`You are a data analysis expert. When working with data:
1. Structure your analysis methodically
2. Identify trends, outliers, and patterns
3. Suggest appropriate statistical methods
4. Recommend visualization approaches
5. Explain findings in plain language`)
}

// NewWebResearchSkill returns a skill for web research.
func NewWebResearchSkill() interfaces.Skill {
	return NewBaseSkill(
		"web_research",
		"Web research and information gathering",
		"web",
	).WithPrompt(`You are a thorough researcher. When conducting research:
1. Break down complex topics into key questions
2. Cross-reference information from multiple sources
3. Distinguish between facts and opinions
4. Cite sources clearly
5. Organize findings in a structured, easy-to-read format`)
}

// NewCreativeWritingSkill returns a skill for creative writing.
func NewCreativeWritingSkill() interfaces.Skill {
	return NewBaseSkill(
		"creative_writing",
		"Creative writing and content generation",
		"content",
	).WithPrompt(`You are a creative writer. When creating content:
1. Use vivid, engaging language
2. Structure content with clear narrative flow
3. Adapt tone and style to the audience
4. Include concrete examples and anecdotes
5. End with a memorable conclusion or call to action`)
}
