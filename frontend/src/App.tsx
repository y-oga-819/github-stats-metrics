import './App.css'
import { PullRequestList } from './ features/pullrequestlist/PullRequestList';
import { BrowserRouter } from 'react-router-dom';
import { Chart } from './ features/pullrequestlist/Chart';
import { useQuery } from "@tanstack/react-query";
import { PR } from "./ features/pullrequestlist/PullRequest";

const fetchPullRequests = async () => {
  const data = await fetch('http://localhost:3000/api/pull_requests')
      .then((res) => res.json())
      .then((data) => {
          console.log(data)
          const result: PR[] = [];
          for (const pr of data) {
              const newPR: PR = {
                  id: pr.Number,
                  title: pr.Title,
                  url: pr.URL,
                  username: pr.Author.Login,
                  iconURL: pr.Author.AvatarURL,
                  repository: pr.Repository.Name,
                  created: new Date(pr.CreatedAt),
                  firstReviewed: new Date(pr.FirstReviewed.Nodes[0].CreatedAt),
                  lastApproved: new Date(pr.LastApprovedAt.Nodes[0].CreatedAt),
                  merged: new Date(pr.MergedAt),
              }
              result.push(newPR)
          }
          return result
      })
      .catch(console.error);
  return data;
};

function App() {
  const {data: pullRequests } = useQuery({queryKey: ['pullRequests'], queryFn: fetchPullRequests});

  return (
    <>
      <Chart prs={pullRequests}/>
      <BrowserRouter>
        <PullRequestList prs={pullRequests}/>
      </BrowserRouter>
    </>
  )
}

export default App
