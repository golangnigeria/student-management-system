package models

// TemplateData holds data sent to templates
type TemplateData struct {
	StringMap map[string]string
	IntMap    map[string]int
	FloatMap  map[string]float64
	Data      map[string]interface{}
	CSRFToken string
	Flash     string
	Warning   string
	Errors    []string
	UserData  *User // User session data
	Path      string
}
