package query

// Template represents a pre-built query template
type Template struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	PrimaryEntity string      `json:"primary_entity"`
	Join          *Join       `json:"join,omitempty"`
	Filters       []Filter    `json:"filters"`
	Parameters    []Parameter `json:"parameters"`
}

// Parameter defines a template parameter
type Parameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // string, array, number
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
}

// templates is the template registry
var templates = map[string]*Template{
	"missing_software": {
		ID:            "missing_software",
		Name:          "Assets Missing Critical Software",
		Description:   "Find assets that do not have a specific software installed (e.g., CrowdStrike, antivirus)",
		PrimaryEntity: "assets",
		Join: &Join{
			Entity: "software_inventory",
			Type:   "left",
			On: JoinCondition{
				Primary: "id",
				Joined:  "asset_id",
			},
		},
		Filters: []Filter{
			{Field: "product_name", Operator: "eq", Value: "{{software_name}}"},
		},
		Parameters: []Parameter{
			{
				Name:        "software_name",
				Type:        "string",
				Description: "Name of the software to check (e.g., 'CrowdStrike Falcon')",
				Default:     "CrowdStrike Falcon",
			},
		},
	},
	"exploitable_cves": {
		ID:            "exploitable_cves",
		Name:          "Assets with Exploitable CVEs",
		Description:   "Find assets that have CVEs with high EPSS scores or in CISA KEV",
		PrimaryEntity: "assets",
		Join: &Join{
			Entity: "findings",
			Type:   "left",
			On: JoinCondition{
				Primary: "id",
				Joined:  "asset_id",
			},
		},
		Filters: []Filter{
			{Field: "epss_score", Operator: "gte", Value: "{{epss_threshold}}"},
		},
		Parameters: []Parameter{
			{
				Name:        "epss_threshold",
				Type:        "number",
				Description: "Minimum EPSS score (0.0-1.0)",
				Default:     0.9,
			},
		},
	},
	"software_vulnerabilities": {
		ID:            "software_vulnerabilities",
		Name:          "Vulnerabilities by Software",
		Description:   "Find findings affecting specific software products",
		PrimaryEntity: "findings",
		Join:          nil, // findings already joined to assets via view
		Filters: []Filter{
			{Field: "cve", Operator: "is_not_null"},
		},
		Parameters: []Parameter{
			{
				Name:        "vendor",
				Type:        "string",
				Description: "Software vendor (e.g., 'Microsoft')",
			},
			{
				Name:        "product",
				Type:        "string",
				Description: "Software product (e.g., 'Windows Server')",
			},
		},
	},
}

// GetTemplate retrieves a template by ID
func GetTemplate(id string) *Template {
	return templates[id]
}

// ListTemplates returns all available templates
func ListTemplates() []*Template {
	result := make([]*Template, 0, len(templates))
	for _, tmpl := range templates {
		result = append(result, tmpl)
	}
	return result
}
