package pull_request

import "github.com/shurcooL/githubv4"

type PullRequest struct {
	Id          githubv4.String
	Number      githubv4.Int
	Title       githubv4.String
	BaseRefName githubv4.String
	HeadRefName githubv4.String
	Author      struct {
		Login     githubv4.String
		AvatarURL githubv4.URI `graphql:"avatarUrl(size:72)"`
	}
	Repository struct {
		Name githubv4.String
	}
	URL           githubv4.URI
	Additions     githubv4.Int
	Deletions     githubv4.Int
	CreatedAt     githubv4.DateTime
	FirstReviewed struct {
		Nodes []struct {
			CreatedAt githubv4.DateTime
		}
	} `graphql:"FirstReviewed: reviews(first: 1)"`
	LastApprovedAt struct {
		Nodes []struct {
			CreatedAt githubv4.DateTime
		}
	} `graphql:"LastApprovedAt: reviews(last: 1, states: APPROVED)"`
	MergedAt githubv4.DateTime
}
