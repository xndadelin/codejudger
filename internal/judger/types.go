package judger

type LanguageConfig struct {
	Compile     string `json:"compile"`
	Extension   string `json:"extension"`
	Run         string `json:"run"`
	File        string `json:"file"`
	Requirement string `json:"requirement,omitempty"`
	Shebang     string `json:"shebang,omitempty"`
}
