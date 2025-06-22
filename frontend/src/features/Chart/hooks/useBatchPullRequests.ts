import { useEffect, useState } from 'react';
import { FetchPullRequests } from '../../pullrequestlist/PullRequestsFetcher';
import { Sprint } from '../../sprintlist/SprintRow';
import { PullRequest } from '../../pullrequestlist/PullRequestRow';

export interface SprintPullRequests {
  sprintId: number;
  pullRequests: PullRequest[];
}

/**
 * バッチ化されたPull Request取得hook
 * N+1クエリ問題を解決するため、全スプリントの期間を統合して1回のAPI呼び出しで取得
 */
export const useBatchPullRequests = (sprintList: Sprint[]) => {
  const [data, setData] = useState<SprintPullRequests[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchBatchPullRequests = async () => {
      if (sprintList.length === 0) return;

      try {
        setLoading(true);
        setError(null);

        // 全スプリントの期間を統合
        const earliestDate = sprintList.reduce((earliest, sprint) => 
          sprint.startDate < earliest ? sprint.startDate : earliest, 
          sprintList[0].startDate
        );
        
        const latestDate = sprintList.reduce((latest, sprint) => 
          sprint.endDate > latest ? sprint.endDate : latest, 
          sprintList[0].endDate
        );

        // 全期間のPull Requestsを1回のAPI呼び出しで取得
        const allPullRequests = await FetchPullRequests({
          id: 0, // バッチ用の仮ID
          startDate: earliestDate,
          endDate: latestDate,
          members: [] // 全メンバーを取得
        });

        // epicブランチを除外
        const filteredPRs = allPullRequests.filter(pr => !pr.branchName.startsWith("epic/"));

        // 各スプリントに該当するPRを分類
        const sprintPullRequests: SprintPullRequests[] = sprintList.map(sprint => {
          const sprintPRs = filteredPRs.filter(pr => {
            return pr.merged >= sprint.startDate && pr.merged <= sprint.endDate;
          });

          return {
            sprintId: sprint.id,
            pullRequests: sprintPRs
          };
        });

        setData(sprintPullRequests);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch pull requests');
        console.error('Error fetching batch pull requests:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchBatchPullRequests();
  }, [sprintList]);

  return { data, loading, error };
};