import { Bar } from "react-chartjs-2";
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

ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend
);

const chartOptions = {
  plugins: {
    title: {
      display: true,
      text: 'スプリントあたりの開発メトリクスチャート',
    },
    annotation: { 
      annotations: [{ 
        type: 'line', // 線を描画
        id: '2dayBorder', 
        mode: 'horizontal', // 線を水平に引く
        scaleID: 'y-axis-0', 
        value: 86400, // 基準となる数値
        borderWidth: 10, // 基準線の太さ
        borderColor: 'red'  // 基準線の色
      }] 
    },  
  },
  responsive: true,
  scales: {
    x: {
      stacked: true,
    },
    y: {
      stacked: true,
    },
  },
}

type Metrics = {
  sprintId: number,
  score: number,
}

export const Chart = () => {
  const sprintList = GetSprintList()
  const labels = sprintList.map((sprint) => sprint.id)

  const [untilFirstReviewedList, setUntilFirstReviewedList] = useState<Metrics[]>([]);
  const [untilLastApprovedList, setUntilLastApprovedList] = useState<Metrics[]>([]);
  const [untilMergedList, setUntilMergedList] = useState<Metrics[]>([]);

  const datasets = {
    labels,
    datasets: [
      {
        label: 'レビューまでにかかった時間',
        data: untilFirstReviewedList.map((metrics) => metrics.score),
        backgroundColor: 'rgb(255, 99, 132)',
      },
      {
        label: '最後のapproveまでにかかった時間',
        data: untilLastApprovedList.map((metrics) => metrics.score),
        backgroundColor: 'rgb(75, 192, 192)',
      },
      {
        label: 'マージまでにかかった時間',
        data: untilMergedList.map((metrics) => metrics.score),
        backgroundColor: 'rgb(53, 162, 235)',
      },
    ],
  };

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
      {datasets
       ? <Bar options={chartOptions} data={datasets} />
       : <div>Loading...</div>
      }
    </>
  )
};