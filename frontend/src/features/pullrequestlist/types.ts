// API レスポンスの型定義
export interface PullRequestAPIResponse {
  ID: string;
  Number: number;
  Title: string;
  BaseRefName: string;
  HeadRefName: string;
  Author: {
    Login: string;
    AvatarURL: string;
  };
  Repository: {
    Name: string;
  };
  URL: string;
  Additions: number;
  Deletions: number;
  CreatedAt: string;
  FirstReviewed: string | null;
  LastApproved: string | null;
  MergedAt: string | null;
}

// クライアント側の Pull Request 型
export interface PullRequest {
  id: number;
  title: string;
  branchName: string;
  url: string;
  username: string;
  iconURL: string;
  repository: string;
  created: Date;
  firstReviewed: Date | null;
  lastApproved: Date | null;
  merged: Date | null;
}

// Fetch エラー型
export interface FetchError {
  message: string;
  status?: number;
  details?: string;
}

// Fetch 結果型
export interface FetchResult<T> {
  data: T;
  error: null;
}

export interface FetchErrorResult {
  data: null;
  error: FetchError;
}

export type FetchPullRequestsResult = FetchResult<PullRequest[]> | FetchErrorResult;

// 型ガード関数
export function isValidPullRequestResponse(data: any): data is PullRequestAPIResponse {
  return (
    data &&
    typeof data.ID === 'string' &&
    typeof data.Number === 'number' &&
    typeof data.Title === 'string' &&
    typeof data.HeadRefName === 'string' &&
    data.Author &&
    typeof data.Author.Login === 'string' &&
    typeof data.Author.AvatarURL === 'string' &&
    data.Repository &&
    typeof data.Repository.Name === 'string' &&
    typeof data.URL === 'string' &&
    typeof data.CreatedAt === 'string'
  );
}

// 日付文字列を安全にDate型に変換
export function safeParseDate(dateString: string | null): Date | null {
  if (!dateString) return null;
  
  const date = new Date(dateString);
  return isNaN(date.getTime()) ? null : date;
}

// APIレスポンスをクライアント型に変換
export function transformPullRequestResponse(apiResponse: PullRequestAPIResponse): PullRequest {
  return {
    id: apiResponse.Number,
    title: apiResponse.Title,
    branchName: apiResponse.HeadRefName,
    url: apiResponse.URL,
    username: apiResponse.Author.Login,
    iconURL: apiResponse.Author.AvatarURL,
    repository: apiResponse.Repository.Name,
    created: new Date(apiResponse.CreatedAt),
    firstReviewed: safeParseDate(apiResponse.FirstReviewed),
    lastApproved: safeParseDate(apiResponse.LastApproved),
    merged: safeParseDate(apiResponse.MergedAt),
  };
}