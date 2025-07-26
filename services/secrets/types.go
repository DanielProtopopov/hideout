package secrets

type (
	TreeNode struct {
		Name     string     `json:"Name"`
		Type     string     `json:"Type"`
		Children []TreeNode `json:"Children,omitempty"`
	}
)
