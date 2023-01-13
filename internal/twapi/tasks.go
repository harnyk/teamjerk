package twapi

type TasksResponse struct {
	Tasks  []Task `json:"todo-items"`
	Status string `json:"STATUS"`
}

type Task struct {
	ID           uint64 `json:"id"`
	Content      string `json:"content"`
	ProjectID    uint64 `json:"project-id"`
	ProjectName  string `json:"project-name"`
	TodoListID   uint64 `json:"todo-list-id"`
	TodoListName string `json:"todo-list-name"`
	CompanyName  string `json:"company-name"`
	CompanyID    uint64 `json:"company-id"`
	TimeIsLogged string `json:"timeIsLogged"`
	CanLogTime   bool   `json:"canLogTime"`
}

type TaskProject struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type TasksGroup struct {
	Project TaskProject
	Tasks   []Task
}

func (tr *TasksResponse) GroupByProject() []TasksGroup {
	groupMap := make(map[uint64][]Task)

	for _, task := range tr.Tasks {
		groupMap[task.ProjectID] = append(groupMap[task.ProjectID], task)
	}

	groups := make([]TasksGroup, 0, len(groupMap))

	for projectID, tasks := range groupMap {
		groups = append(groups, TasksGroup{
			Project: TaskProject{
				ID:   projectID,
				Name: tasks[0].ProjectName,
			},
			Tasks: tasks,
		})
	}

	return groups
}
