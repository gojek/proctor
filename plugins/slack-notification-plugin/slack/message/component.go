package message

type textComponent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type sectionComponent struct {
	Type   string          `json:"type"`
	Text   textComponent   `json:"text"`
	Fields []textComponent `json:"fields"`
}
