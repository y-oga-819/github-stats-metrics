import { useMemo } from "react";
import { Bar } from "react-chartjs-2";
import { Sprint } from "../sprintlist/SprintRow";
import { Metrics } from "./Chart";

const chartOptions = {
    plugins: {
      title: {
        display: true,
        text: 'スプリントあたりの開発メトリクス',
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

type MetricsChartProps = {
    sprintList: Sprint[]
    untilFirstReviewedList: Metrics[]
    untilLastApprovedList: Metrics[]
    untilMergedList: Metrics[]
}

export const MetricsChart: React.FC<MetricsChartProps> = ({sprintList, untilFirstReviewedList, untilLastApprovedList, untilMergedList}) => {
    const datasets = useMemo(() => {
        const labels = sprintList.map((sprint) => sprint.id);

        const processMetrics = (metrics: Metrics[]) => {
            return Object.values(
                metrics
                    .sort((a, b) => a.sprintId - b.sprintId)
                    .reduce((prev, current) => ({[current.sprintId]: current.score, ...prev}), {})
            );
        };

        return {
            labels,
            datasets: [
                {
                    label: 'レビューまでにかかった時間',
                    data: processMetrics(untilFirstReviewedList),
                    backgroundColor: 'rgb(255, 99, 132)',
                },
                {
                    label: '最後のapproveまでにかかった時間',
                    data: processMetrics(untilLastApprovedList),
                    backgroundColor: 'rgb(75, 192, 192)',
                },
                {
                    label: 'マージまでにかかった時間',
                    data: processMetrics(untilMergedList),
                    backgroundColor: 'rgb(53, 162, 235)',
                },
            ],
        };
    }, [sprintList, untilFirstReviewedList, untilLastApprovedList, untilMergedList]);

    return (
        <Bar options={chartOptions} data={datasets} />
    );
}