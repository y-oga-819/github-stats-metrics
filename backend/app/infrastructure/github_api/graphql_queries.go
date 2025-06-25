package github_api

import (
	"github.com/shurcooL/githubv4"
)

// ExtendedGraphQLQuery は拡張されたGraphQLクエリ構造体
type ExtendedGraphQLQuery struct {
	Search struct {
		CodeCount githubv4.Int
		PageInfo  struct {
			HasNextPage githubv4.Boolean
			EndCursor   githubv4.String
		}
		Nodes []struct {
			Pr ExtendedPullRequest `graphql:"... on PullRequest"`
		}
	} `graphql:"search(type: $searchType, first: 100, after: $cursor, query: $query)"`
	RateLimit struct {
		Cost      githubv4.Int
		Limit     githubv4.Int
		Remaining githubv4.Int
		ResetAt   githubv4.DateTime
	}
}

// ExtendedPullRequest はPRの詳細データを取得するための拡張構造体
type ExtendedPullRequest struct {
	// 基本情報
	Id          githubv4.String
	Number      githubv4.Int
	Title       githubv4.String
	BaseRefName githubv4.String
	HeadRefName githubv4.String
	URL         githubv4.URI
	CreatedAt   githubv4.DateTime
	MergedAt    githubv4.DateTime
	
	// 作者情報
	Author struct {
		Login     githubv4.String
		AvatarURL githubv4.URI `graphql:"avatarUrl(size:72)"`
	}
	
	// リポジトリ情報
	Repository struct {
		Name githubv4.String
	}
	
	// サイズメトリクス
	Additions   githubv4.Int
	Deletions   githubv4.Int
	ChangedFiles githubv4.Int
	
	// ファイル変更詳細
	Files struct {
		Nodes []struct {
			Path      githubv4.String
			Additions githubv4.Int
			Deletions githubv4.Int
			ChangeType githubv4.String // ADDED, DELETED, MODIFIED, RENAMED
		}
	} `graphql:"files(first: 100)"`
	
	// レビュー情報（詳細）
	Reviews struct {
		TotalCount githubv4.Int
		Nodes []struct {
			Id        githubv4.String
			CreatedAt githubv4.DateTime
			State     githubv4.PullRequestReviewState
			Author struct {
				Login githubv4.String
			}
			Comments struct {
				TotalCount githubv4.Int
			}
		}
	} `graphql:"reviews(first: 50)"`
	
	// レビューコメント（詳細）
	ReviewComments struct {
		TotalCount githubv4.Int
		Nodes []struct {
			Id        githubv4.String
			CreatedAt githubv4.DateTime
			Author struct {
				Login githubv4.String
			}
			Path     githubv4.String
			Position githubv4.Int
		}
	} `graphql:"reviewComments(first: 100)"`
	
	// コミット情報
	Commits struct {
		TotalCount githubv4.Int
		Nodes []struct {
			Commit struct {
				MessageHeadline githubv4.String
				CommittedDate   githubv4.DateTime
				Author struct {
					Name githubv4.String
					Date githubv4.DateTime
				}
			}
		}
	} `graphql:"commits(first: 50)"`
	
	// レビュー要求
	ReviewRequests struct {
		TotalCount githubv4.Int
		Nodes []struct {
			RequestedReviewer struct {
				User struct {
					Login githubv4.String
				} `graphql:"... on User"`
			}
		}
	} `graphql:"reviewRequests(first: 20)"`
	
	// タイムライン情報（レビュー要求、承認等のイベント）
	TimelineItems struct {
		Nodes []struct {
			ReviewRequestedEvent struct {
				CreatedAt githubv4.DateTime
				RequestedReviewer struct {
					User struct {
						Login githubv4.String
					} `graphql:"... on User"`
				}
			} `graphql:"... on ReviewRequestedEvent"`
			
			PullRequestReview struct {
				CreatedAt githubv4.DateTime
				State     githubv4.PullRequestReviewState
				Author struct {
					Login githubv4.String
				}
			} `graphql:"... on PullRequestReview"`
			
			PullRequestReviewComment struct {
				CreatedAt githubv4.DateTime
				Author struct {
					Login githubv4.String
				}
			} `graphql:"... on PullRequestReviewComment"`
		}
	} `graphql:"timelineItems(first: 100, itemTypes: [PULL_REQUEST_REVIEW, REVIEW_REQUESTED_EVENT, PULL_REQUEST_REVIEW_COMMENT])"`
}

// ReviewMetricsQuery はレビューメトリクス専用のクエリ
type ReviewMetricsQuery struct {
	Node struct {
		PullRequest ExtendedPullRequest `graphql:"... on PullRequest"`
	} `graphql:"node(id: $prId)"`
	RateLimit struct {
		Cost      githubv4.Int
		Limit     githubv4.Int
		Remaining githubv4.Int
		ResetAt   githubv4.DateTime
	}
}

// FileDetailsQuery はファイル詳細取得専用のクエリ
type FileDetailsQuery struct {
	Node struct {
		PullRequest struct {
			Files struct {
				PageInfo struct {
					HasNextPage githubv4.Boolean
					EndCursor   githubv4.String
				}
				Nodes []struct {
					Path       githubv4.String
					Additions  githubv4.Int
					Deletions  githubv4.Int
					ChangeType githubv4.String
					Patch      githubv4.String
				}
			} `graphql:"files(first: 100, after: $cursor)"`
		} `graphql:"... on PullRequest"`
	} `graphql:"node(id: $prId)"`
}

// CommitDetailsQuery はコミット詳細取得専用のクエリ
type CommitDetailsQuery struct {
	Node struct {
		PullRequest struct {
			Commits struct {
				PageInfo struct {
					HasNextPage githubv4.Boolean
					EndCursor   githubv4.String
				}
				Nodes []struct {
					Commit struct {
						Oid             githubv4.String
						MessageHeadline githubv4.String
						Message         githubv4.String
						CommittedDate   githubv4.DateTime
						Author struct {
							Name  githubv4.String
							Email githubv4.String
							Date  githubv4.DateTime
						}
						Additions githubv4.Int
						Deletions githubv4.Int
						ChangedFiles githubv4.Int
					}
				}
			} `graphql:"commits(first: 100, after: $cursor)"`
		} `graphql:"... on PullRequest"`
	} `graphql:"node(id: $prId)"`
}

// ReviewTimelineQuery はレビュータイムライン専用のクエリ
type ReviewTimelineQuery struct {
	Node struct {
		PullRequest struct {
			TimelineItems struct {
				PageInfo struct {
					HasNextPage githubv4.Boolean
					EndCursor   githubv4.String
				}
				Nodes []struct {
					// レビュー要求イベント
					ReviewRequestedEvent struct {
						CreatedAt githubv4.DateTime
						Actor struct {
							Login githubv4.String
						}
						RequestedReviewer struct {
							User struct {
								Login githubv4.String
							} `graphql:"... on User"`
						}
					} `graphql:"... on ReviewRequestedEvent"`
					
					// レビュー要求削除イベント
					ReviewRequestRemovedEvent struct {
						CreatedAt githubv4.DateTime
						Actor struct {
							Login githubv4.String
						}
						RequestedReviewer struct {
							User struct {
								Login githubv4.String
							} `graphql:"... on User"`
						}
					} `graphql:"... on ReviewRequestRemovedEvent"`
					
					// レビューイベント
					PullRequestReview struct {
						Id        githubv4.String
						CreatedAt githubv4.DateTime
						State     githubv4.PullRequestReviewState
						Author struct {
							Login githubv4.String
						}
						SubmittedAt githubv4.DateTime
					} `graphql:"... on PullRequestReview"`
					
					// レビューコメントイベント
					PullRequestReviewComment struct {
						Id        githubv4.String
						CreatedAt githubv4.DateTime
						Author struct {
							Login githubv4.String
						}
						Path     githubv4.String
						Position githubv4.Int
					} `graphql:"... on PullRequestReviewComment"`
					
					// PR準備完了イベント
					ReadyForReviewEvent struct {
						CreatedAt githubv4.DateTime
						Actor struct {
							Login githubv4.String
						}
					} `graphql:"... on ReadyForReviewEvent"`
					
					// マージイベント
					MergedEvent struct {
						CreatedAt githubv4.DateTime
						Actor struct {
							Login githubv4.String
						}
						MergeRefName githubv4.String
					} `graphql:"... on MergedEvent"`
				}
			} `graphql:"timelineItems(first: 100, after: $cursor, itemTypes: [PULL_REQUEST_REVIEW, REVIEW_REQUESTED_EVENT, REVIEW_REQUEST_REMOVED_EVENT, PULL_REQUEST_REVIEW_COMMENT, READY_FOR_REVIEW_EVENT, MERGED_EVENT])"`
		} `graphql:"... on PullRequest"`
	} `graphql:"node(id: $prId)"`
}

// getLimitedQuery は制限されたフィールドのみを取得するクエリ（レート制限対策）
func getLimitedQuery() interface{} {
	return &struct {
		Search struct {
			CodeCount githubv4.Int
			PageInfo  struct {
				HasNextPage githubv4.Boolean
				EndCursor   githubv4.String
			}
			Nodes []struct {
				Pr struct {
					Id          githubv4.String
					Number      githubv4.Int
					Title       githubv4.String
					BaseRefName githubv4.String
					HeadRefName githubv4.String
					URL         githubv4.URI
					CreatedAt   githubv4.DateTime
					MergedAt    githubv4.DateTime
					Additions   githubv4.Int
					Deletions   githubv4.Int
					Author struct {
						Login githubv4.String
					}
					Repository struct {
						Name githubv4.String
					}
				} `graphql:"... on PullRequest"`
			}
		} `graphql:"search(type: $searchType, first: 100, after: $cursor, query: $query)"`
		RateLimit struct {
			Cost      githubv4.Int
			Limit     githubv4.Int
			Remaining githubv4.Int
			ResetAt   githubv4.DateTime
		}
	}{}
}

// getDetailedQuery は詳細なフィールドを取得するクエリ（個別PR取得用）
func getDetailedQuery() interface{} {
	return &ExtendedGraphQLQuery{}
}