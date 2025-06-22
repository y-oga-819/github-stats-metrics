package pull_request

import (
	"fmt"
	"time"
)

type GetPullRequestsRequest struct {
	StartDate  string   `schema:"startdate,required"`
	EndDate    string   `schema:"enddate,required"`
	Developers []string `schema:"developers,required"`
}

// バリデーションロジック
func (req GetPullRequestsRequest) Validate() error {
	if req.StartDate == "" {
		return fmt.Errorf("start date is required")
	}
	if req.EndDate == "" {
		return fmt.Errorf("end date is required")
	}
	if len(req.Developers) == 0 {
		return fmt.Errorf("at least one developer is required")
	}
	
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return fmt.Errorf("invalid start date format: %w", err)
	}
	
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return fmt.Errorf("invalid end date format: %w", err)
	}
	
	if startDate.After(endDate) {
		return fmt.Errorf("start date must be before end date")
	}
	
	return nil
}

func (req GetPullRequestsRequest) GetStartDate() (time.Time, error) {
	return time.Parse("2006-01-02", req.StartDate)
}

func (req GetPullRequestsRequest) GetEndDate() (time.Time, error) {
	return time.Parse("2006-01-02", req.EndDate)
}
