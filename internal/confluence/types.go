package confluence

// Page represents a Confluence page
type Page struct {
	ID        string         `json:"id,omitempty"`
	Type      string         `json:"type"`
	Title     string         `json:"title"`
	Space     Space          `json:"space"`
	Body      Body           `json:"body"`
	Version   *Version       `json:"version,omitempty"`
	Ancestors []PageAncestor `json:"ancestors,omitempty"`
}

// PageAncestor represents a parent page
type PageAncestor struct {
	ID string `json:"id"`
}

// Space represents a Confluence space
type Space struct {
	Key string `json:"key"`
}

// Body represents page content
type Body struct {
	Storage Storage `json:"storage"`
}

// Storage represents page storage format
type Storage struct {
	Value          string `json:"value"`
	Representation string `json:"representation"`
}

// Version represents page version
type Version struct {
	Number int `json:"number"`
}

// SearchResponse represents a page search response
type SearchResponse struct {
	Results []Page `json:"results"`
}
