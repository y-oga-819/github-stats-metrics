import { Bar } from "react-chartjs-2";
import { Sprint } from "../sprintlist/SprintRow";
import { Metrics } from "./Chart";

const chartOptions = {
    plugins: {
      title: {
        display: true,
        text: 'スプリントあたりのPR数',
      },
    },
    responsive: true,
  }

type DevDayDeveloperChartProps = {
    sprintList: Sprint[]
    devDayDeveloperList: Metrics[]
}

export const DevDayDeveloperList: React.FC<DevDayDeveloperChartProps> = ({sprintList, devDayDeveloperList}) => {
    const labels = sprintList.map((sprint) => sprint.id)
    const datasets = {
        labels,
        datasets: [
          {
            label: 'Dev / Day / Developer',
            data: devDayDeveloperList.map((metrics) => metrics.score ),
            backgroundColor: 'rgb(75, 192, 192)',
          },
        ],
      };

    return (
        <Bar options={chartOptions} data={datasets} />
    );
}