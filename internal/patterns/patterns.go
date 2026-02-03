package patterns

type Pattern struct {
	Regex string
}

var BuiltinPatterns = map[string]Pattern{
	"claude": {Regex: `thinking`},
	"codex":  {Regex: `esc to interrupt`},
	"gemini": {Regex: `esc to cancel`},
}

func GetPattern(toolName string) *Pattern {
	if p, ok := BuiltinPatterns[toolName]; ok {
		return &p
	}
	return nil
}

func DefaultPattern() *Pattern {
	return &Pattern{
		Regex: `esc to interrupt`,
	}
}
