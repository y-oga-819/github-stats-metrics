import { Bar } from "react-chartjs-2";
import { Sprint } from "../sprintlist/SprintRow";
import { Metrics } from "./Chart";

const chartOptions = {
    plugins: {
      title: {
        display: true,
        text: 'スプリントあたりのPR数',
      },
      annotation: { 
        annotations: [{
          id: 'hline', 
          type: 'line', // 線を描画
          mode: 'horizontal', // 線を水平に引く
          scaleID: 'y-axis-0', 
          value: 20, // 基準となる数値
          borderWidth: 10, // 基準線の太さ
          borderColor: 'red'  // 基準線の色
        }] 
      },
      legend: {
        position: 'bottom',
      }
    },
    responsive: true,
  }

type PrCountChartProps = {
    sprintList: Sprint[]
    prCountList: Metrics[]
}

export const PrCountChart: React.FC<PrCountChartProps> = ({sprintList, prCountList}) => {
    const labels = sprintList.map((sprint) => sprint.id)
    const datasets = {
        labels,
        datasets: [
          {
            label: 'マージしたPR数',
            data: prCountList.map((metrics) => metrics.score),
            backgroundColor: 'rgb(255, 99, 132)',
          },
        ],
      };

    return (
        <Bar options={chartOptions} data={datasets} />
    );
}