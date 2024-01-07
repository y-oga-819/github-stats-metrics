package pull_request

type GetPullRequestsRequest struct {
	StartDate  string   `schema:startdate,required`
	EndDate    string   `schema:enddate,required`
	Developers []string `schema:developers,required`
}
