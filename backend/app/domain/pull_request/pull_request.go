package pull_request

import (
	developerDomain "github-stats-metrics/domain/developer"
	"time"
)

type PullRequest struct {
	Id            string
	Title         string
	Status        string
	Author        developerDomain.Developer
	Opened        time.Time
	FirstReviewed time.Time
	Approved      time.Time
	Merged        time.Time
}
