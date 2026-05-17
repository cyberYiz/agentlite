// Package calculator provides a simple arithmetic tool.
package calculator

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/cyberYiz/agent-sdk/pkg/interfaces"
)

// Tool implements interfaces.Tool for basic arithmetic.
type Tool struct{}

// New creates a new calculator tool.
func New() *Tool { return &Tool{} }

// Ensure Tool implements interfaces.Tool
var _ interfaces.Tool = (*Tool)(nil)

func (t *Tool) Name() string { return "calculator" }
func (t *Tool) Description() string {
	return "Performs basic arithmetic: addition, subtraction, multiplication, division. Input expression like '2 + 3 * 4' (supports +, -, *, /)."
}
func (t *Tool) DisplayName() string { return "Calculator" }

func (t *Tool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"expression": map[string]interface{}{
				"type":        "string",
				"description": "The arithmetic expression to evaluate (e.g., '2 + 3 * 4')",
			},
		},
		"required": []string{"expression"},
	}
}

func (t *Tool) Execute(ctx context.Context, args string) (string, error) {
	// Context-aware: abort early if cancelled
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("calculator: %w", err)
	}

	// Simple expression parsing: split by spaces and evaluate left-to-right
	// This is intentionally simple; in production use a proper expression parser.
	expr := extractField(args, "expression")
	if expr == "" {
		expr = args
	}
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return "", fmt.Errorf("empty expression")
	}

	result, err := evaluate(expr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s = %g", expr, result), nil
}

// extractField tries to extract a JSON field; falls back to raw input.
func extractField(args string, key string) string {
	args = strings.TrimSpace(args)
	// Simple JSON extraction
	if strings.HasPrefix(args, "{") {
		// Try to find "key": "value" or "key": value
		search := fmt.Sprintf(`"%s":`, key)
		idx := strings.Index(args, search)
		if idx >= 0 {
			rest := args[idx+len(search):]
			rest = strings.TrimSpace(rest)
			if strings.HasPrefix(rest, `"`) {
				rest = rest[1:]
				end := strings.Index(rest, `"`)
				if end >= 0 {
					return rest[:end]
				}
			}
		}
	}
	return args
}

func evaluate(expr string) (float64, error) {
	tokens := strings.Fields(expr)
	if len(tokens) == 0 {
		return 0, fmt.Errorf("no tokens")
	}

	// Parse numbers and operators
	var nums []float64
	var ops []string

	for _, tok := range tokens {
		switch tok {
		case "+", "-", "*", "/":
			ops = append(ops, tok)
		default:
			n, err := strconv.ParseFloat(tok, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid number: %s", tok)
			}
			nums = append(nums, n)
		}
	}

	if len(nums) == 0 {
		return 0, fmt.Errorf("no numbers found")
	}
	if len(nums) != len(ops)+1 {
		return 0, fmt.Errorf("malformed expression: %d numbers, %d operators", len(nums), len(ops))
	}

	// Apply * and / first (left to right)
	i := 0
	for i < len(ops) {
		if ops[i] == "*" || ops[i] == "/" {
			op := ops[i]
			a, b := nums[i], nums[i+1]
			var res float64
			switch op {
			case "*":
				res = a * b
			case "/":
				if b == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				res = a / b
			}
			nums[i] = res
			nums = append(nums[:i+1], nums[i+2:]...)
			ops = append(ops[:i], ops[i+1:]...)
		} else {
			i++
		}
	}

	// Apply + and -
	result := nums[0]
	for i, op := range ops {
		switch op {
		case "+":
			result += nums[i+1]
		case "-":
			result -= nums[i+1]
		}
	}

	return result, nil
}
