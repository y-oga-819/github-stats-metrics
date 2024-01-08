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
import { FetchPullRequests } from "../pullrequestlist/PullRequestsFetcher";
import { MetricsChart } from "./MetricsChart";

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
  const sprintList = GetSprintList()

  const [untilFirstReviewedList, setUntilFirstReviewedList] = useState<Metrics[]>([]);
  const [untilLastApprovedList, setUntilLastApprovedList] = useState<Metrics[]>([]);
  const [untilMergedList, setUntilMergedList] = useState<Metrics[]>([]);

  useEffect(() => {
    // スプリント毎のPRをチャートに反映する
    sprintList.map((sprint) => {
      // 1スプリント分のPRを取得
      const prs = FetchPullRequests(sprint)
      prs.then((prs) => prs.filter((pr) => !pr.branchName.startsWith("epic/"))) // epicブランチは除外する
      .then((prs) => {
        // PR数を計算
        const prCount = prs.length

        // レビューまでにかかった時間を計算
        const untilFirstReviewed = prs
          .map((pr) => (pr.firstReviewed.getTime() - pr.created.getTime()) / 1000)
          .reduce((a, b) => a + b, 0) / prCount;

        // 最後のapproveまでにかかった時間を計算
        const untilLastApproved = prs
          .map((pr) => (pr.lastApproved.getTime() - pr.firstReviewed.getTime()) / 1000)
          .reduce((a, b) => a + b, 0) / prCount  
  
        // マージまでにかかった時間を計算
        const untilMerged = prs
          .map((pr) => (pr.merged.getTime() - pr.lastApproved.getTime()) / 1000)
          .reduce((a, b) => a + b, 0) / prCount

        // これらの値をチャート用のデータに追加する
        setUntilFirstReviewedList((untilFirstReviewedList) => [...untilFirstReviewedList, {sprintId: sprint.id, score: untilFirstReviewed}])
        setUntilLastApprovedList((untilLastApprovedList) => [...untilLastApprovedList, {sprintId: sprint.id, score: untilLastApproved}])
        setUntilMergedList((untilMergedList) => [...untilMergedList, {sprintId: sprint.id, score: untilMerged}])
      })
    })
}, []);

return (
    <>
      <MetricsChart sprintList={sprintList} untilFirstReviewedList={untilFirstReviewedList} untilLastApprovedList={untilLastApprovedList} untilMergedList={untilMergedList} />
    </>
  )
};