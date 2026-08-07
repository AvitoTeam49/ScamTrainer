package scenario

type Role string

const (
	RoleBuyer  Role = "buyer"
	RoleSeller Role = "seller"
)

type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

type NodeType string

const (
	NodeTypeDecision NodeType = "decision"
	NodeTypeEnding   NodeType = "ending"
)

type Scenario struct {
	ID          string           `yaml:"id"`
	Title       string           `yaml:"title"`
	Role        Role             `yaml:"role"`
	Difficulty  Difficulty       `yaml:"difficulty"`
	StartNodeID string           `yaml:"start_node_id"`
	LLM         ScenarioLLM      `yaml:"llm"`
	Nodes       map[string]*Node `yaml:"nodes"`
}

type ScenarioLLM struct {
	CharacterPrompt string `yaml:"character_prompt"`
}

type Node struct {
	ID          string       `yaml:"-"`
	Type        NodeType     `yaml:"type"`
	Message     Message      `yaml:"message"`
	LLM         NodeLLM      `yaml:"llm"`
	Transitions []Transition `yaml:"transitions"`
	Outcome     string       `yaml:"outcome"`
	Title       string       `yaml:"title"`
}

type NodeLLM struct {
	ReplyInstruction string `yaml:"reply_instruction"`
}

type Message struct {
	Author string `yaml:"author"`
	Text   string `yaml:"text"`
}

type Transition struct {
	ID          string   `yaml:"id"`
	Description string   `yaml:"description"`
	Examples    []string `yaml:"examples"`
	ToNodeID    string   `yaml:"to_node_id"`
	ScoreDelta  int      `yaml:"score_delta"`
	Feedback    string   `yaml:"feedback"`
	RiskTags    []string `yaml:"risk_tags"`
}
