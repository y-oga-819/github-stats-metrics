import { Member, Sprint } from "../sprintlist/SprintRow";
import { PR } from "./PullRequest";
import { 
  PullRequestAPIResponse, 
  PullRequest as TypedPullRequest,
  FetchPullRequestsResult,
  isValidPullRequestResponse,
  transformPullRequestResponse 
} from "./types";

/**
 * 型安全なPull Requests取得関数
 * 包括的なエラーハンドリングとnull安全性を提供
 */
export const FetchPullRequests = async (sprint: Sprint): Promise<PR[]> => {
    try {
        const result = await fetchPullRequestsWithTypes(sprint);
        
        if (result.error) {
            console.error('Pull requests fetch failed:', result.error);
            return [];
        }
        
        // 旧型との互換性のため変換
        return result.data.map(convertToLegacyPR).filter(Boolean) as PR[];
    } catch (error) {
        console.error('Unexpected error in FetchPullRequests:', error);
        return [];
    }
};

/**
 * 完全型安全なPull Request取得（新しいAPI）
 */
export const fetchPullRequestsWithTypes = async (sprint: Sprint): Promise<FetchPullRequestsResult> => {
    try {
        // URLパラメータの構築
        const params = new URLSearchParams({
            startdate: sprint.startDate.toISOString().split('T')[0],
            enddate: sprint.endDate.toISOString().split('T')[0],
        });
        
        // 開発者パラメータを追加（型安全）
        sprint.members.forEach((member: Member) => {
            if (member.name && typeof member.name === 'string') {
                params.append('developers', member.name);
            }
        });

        const response = await fetch(`http://localhost:8080/api/pull_requests?${params}`);
        
        // HTTPステータスチェック
        if (!response.ok) {
            return {
                data: null,
                error: {
                    message: `HTTP Error: ${response.status} ${response.statusText}`,
                    status: response.status
                }
            };
        }

        // Content-Typeチェック
        const contentType = response.headers.get('content-type');
        if (!contentType?.includes('application/json')) {
            return {
                data: null,
                error: {
                    message: 'Invalid response: Expected JSON content-type',
                    details: `Received: ${contentType}`
                }
            };
        }

        const rawData = await response.json();
        
        // レスポンスが配列かチェック
        if (!Array.isArray(rawData)) {
            return {
                data: null,
                error: {
                    message: 'Invalid response: Expected array of pull requests',
                    details: `Received: ${typeof rawData}`
                }
            };
        }

        const validPullRequests: TypedPullRequest[] = [];
        const errors: string[] = [];

        // 各Pull Requestを型安全に処理
        for (let i = 0; i < rawData.length; i++) {
            const item = rawData[i];
            
            if (!isValidPullRequestResponse(item)) {
                errors.push(`Invalid pull request at index ${i}: missing required fields`);
                continue;
            }

            try {
                const pullRequest = transformPullRequestResponse(item);
                validPullRequests.push(pullRequest);
            } catch (transformError) {
                errors.push(`Failed to transform pull request at index ${i}: ${transformError}`);
            }
        }

        // 変換エラーがある場合はログ出力
        if (errors.length > 0) {
            console.warn('Pull request transformation errors:', errors);
        }

        return {
            data: validPullRequests,
            error: null
        };

    } catch (networkError) {
        return {
            data: null,
            error: {
                message: 'Network or parsing error occurred',
                details: networkError instanceof Error ? networkError.message : String(networkError)
            }
        };
    }
};

/**
 * 新しい型から旧型への変換（後方互換性）
 */
function convertToLegacyPR(pr: TypedPullRequest): PR | null {
    // 必須フィールドのnullチェック
    if (!pr.firstReviewed || !pr.lastApproved || !pr.merged) {
        console.warn(`Skipping PR ${pr.id}: missing required dates`);
        return null;
    }

    return {
        id: pr.id,
        title: pr.title,
        branchName: pr.branchName,
        url: pr.url,
        username: pr.username,
        iconURL: pr.iconURL,
        repository: pr.repository,
        created: pr.created,
        firstReviewed: pr.firstReviewed,
        lastApproved: pr.lastApproved,
        merged: pr.merged,
    };
}