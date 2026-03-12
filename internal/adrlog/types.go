package adrlog

type Repository struct {
	CWD    string
	ADRDir string
}

type Record struct {
	Number   int
	Filename string
	Path     string
	Title    string
	Date     string
	Status   []string
	Content  string
}

type ValidationIssue struct {
	Path     string
	Severity string
	Message  string
}
