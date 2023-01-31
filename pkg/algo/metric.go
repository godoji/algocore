package algo

type Metrics struct {
	Elapsed      int64 `json:"elapsed"`
	StartTime    int64 `json:"startTime"`
	TotalBlocks  int   `json:"totalBlocks"`
	Progress     int   `json:"progress"`
	Finished     bool  `json:"finished"`
	CurrentBlock int   `json:"currentBlock"`
	Running      bool  `json:"running"`
}
