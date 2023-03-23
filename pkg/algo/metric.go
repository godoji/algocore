package algo

type Status struct {
	Elapsed   int64 `json:"elapsed"`
	StartTime int64 `json:"startTime"`
	Finished  bool  `json:"finished"`
	Running   bool  `json:"running"`
}
