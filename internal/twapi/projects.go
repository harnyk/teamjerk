package twapi

type ProjectsResponse struct {
	Projects []Project `json:"projects"`
	Status   string    `json:"STATUS"`
}

type Project struct {
	Category    ProjectCategory `json:"category"`
	Company     ProjectCompany  `json:"company"`
	Description string          `json:"description"`
	ID          string          `json:"id"`
	IsBillable  bool            `json:"isBillable"`
	Name        string          `json:"name"`
}

type ProjectCompany struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProjectCategory struct {
	Color string `json:"color"`
	ID    string `json:"id"`
	Name  string `json:"name"`
}
