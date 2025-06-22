import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js';
import { GetSprintList } from "../sprintlist/GetConstSprintList";
import { useEffect, useState } from "react";
import { MetricsChart } from "./MetricsChart";
import { PrCountChart } from './PrCountChart';
import { DevDayDeveloperList } from './DevDayDeveloperChart';
import { useBatchPullRequests } from './hooks/useBatchPullRequests';
import { MetricsCalculator } from './utils/metricsCalculator';

ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend
);

export type Metrics = {
  sprintId: number,
  score: number,
}

export const Chart = () => {
  const sprintList = GetSprintList();
  
  // バッチ化されたPull Requests取得（N+1クエリ問題解決）
  const { data: sprintPullRequests, loading, error } = useBatchPullRequests(sprintList);

  const [untilFirstReviewedList, setUntilFirstReviewedList] = useState<Metrics[]>([]);
  const [untilLastApprovedList, setUntilLastApprovedList] = useState<Metrics[]>([]);
  const [untilMergedList, setUntilMergedList] = useState<Metrics[]>([]);
  const [prCountList, setPrCountList] = useState<Metrics[]>([]);
  const [devDayDeveloperList, setDevDayDeveloperList] = useState<Metrics[]>([]);

  useEffect(() => {
    if (loading || error || sprintPullRequests.length === 0) return;

    // バッチで取得したデータから各スプリントのメトリクスを一括計算
    const allMetrics = sprintPullRequests.map(({ sprintId, pullRequests }) => {
      const sprint = sprintList.find(s => s.id === sprintId);
      if (!sprint) return null;

      return MetricsCalculator.calculateSprintMetrics(sprintId, pullRequests, sprint);
    }).filter(Boolean);

    // 計算結果を各状態に設定（1回のバッチ処理）
    setUntilFirstReviewedList(allMetrics.map(m => m!.untilFirstReviewed));
    setUntilLastApprovedList(allMetrics.map(m => m!.untilLastApproved));
    setUntilMergedList(allMetrics.map(m => m!.untilMerged));
    setPrCountList(allMetrics.map(m => m!.prCount));
    setDevDayDeveloperList(allMetrics.map(m => m!.devDayDeveloper));
  }, [sprintPullRequests, loading, error, sprintList]);

  // ローディング中の表示
  if (loading) {
    return <div>Loading metrics...</div>;
  }

  // エラー時の表示
  if (error) {
    return <div>Error loading metrics: {error}</div>;
  }

  return (
    <>
      <MetricsChart 
        sprintList={sprintList} 
        untilFirstReviewedList={untilFirstReviewedList} 
        untilLastApprovedList={untilLastApprovedList} 
        untilMergedList={untilMergedList} 
      />
      <PrCountChart sprintList={sprintList} prCountList={prCountList} />
      <DevDayDeveloperList sprintList={sprintList} devDayDeveloperList={devDayDeveloperList} />
    </>
  );
};