import { ApexOptions } from "apexcharts";
import { PR } from "./PullRequest";

export const Chart = () => {
  // const prIds = prs?.map((pr: PR) => {return pr.id})
  //  const betweenFirstReviewAndCreated = prs?.map((pr: PR) => {
  //   return pr.firstReviewed.getTime() - pr.created.getTime()
  // }) ?? []
  // const betweenLastApprovedAndFirstReviewed = prs?.map((pr: PR) => {
  //   return pr.lastApproved.getTime() - pr.firstReviewed.getTime()
  // }) ?? []
  // const betweenMergedAndLastApproved = prs?.map((pr: PR) => {
  //   return pr.merged.getTime() - pr.lastApproved.getTime()
  // }) ?? []

  // const chartData: ApexOptions = {
  //     chart: {
  //       type: "line",
  //       id: "apexchart-example",
  //       stacked: true,
  //     },
  //     xaxis: {
  //       categories: prIds
  //     },
  //     fill: {
  //       type: "gradient",
  //       gradient: {
  //         shade: "light",
  //         type: "horizontal",
  //         shadeIntensity: 0.5,
  //         gradientToColors: undefined, // optional, if not defined - uses the shades of same color in series
  //         inverseColors: true,
  //         opacityFrom: 1,
  //         opacityTo: 1,
  //         stops: [0, 50, 100]
  //       }
  //     },
  //     legend: {
  //       // position: '',
  //       width: 300
  //       // position: 'top',
  //     },
  //     series: [
  //       {
  //         name: "betweenFirstReviewAndCreated",
  //         type: "column",
  //         data: betweenFirstReviewAndCreated,
  //       },
  //       {
  //           name: "betweenLastApprovedAndFirstReviewed",
  //           type: "column",
  //           data: betweenLastApprovedAndFirstReviewed,
  //       },
  //       {
  //         name: "betweenMergedAndLastApproved",
  //         type: "column",
  //         data: betweenMergedAndLastApproved,
  //       }, 
  //       {
  //         name: "Time Traveled",
  //         type: "line",
  //         data: [23, 42, 35, 27, 43, 22, 17, 31, 42, 22, 12, 16],
  //       }
  //     ]
  //   };
  
    return (
        // <ReactApexChart options={chartData} series={chartData.series} />
        <h1>Chart Page</h1>
    );
};