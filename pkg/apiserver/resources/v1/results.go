package v1

type ResourcesSummaryResult struct {
	Clusters ResourcesSummaryClusterResult `json:"clusters"`
	Nodes    ResourcesSummaryNodeResult    `json:"nodes"`
}

type ResourcesSummaryClusterResult struct {
	NotInstall   int `json:"notInstall"`
	Initializing int `json:"initializing"`
	Running      int `json:"running"`
	Upgrading    int `json:"upgrading"`
	Deleting     int `json:"deleting"`
	Deleted      int `json:"deleted"`
	Failed       int `json:"failed"`
}

type ResourcesSummaryNodeResult struct {
	Creating      int `json:"creating"`
	Warning       int `json:"warning"`
	Ready         int `json:"ready"`
	Running       int `json:"running"`
	Fault         int `json:"fault"`
	Unschedulable int `json:"unschedulable"`
}
