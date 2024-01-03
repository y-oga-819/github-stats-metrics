import './App.css'
import { PullRequestList } from './features/pullrequestlist/PullRequestList';
import { BrowserRouter } from 'react-router-dom';
import { SprintList } from './features/sprintlist/SprintList';

export const App = () => {
  return (
    <>
      {/* <Chart prs={pullRequests}/> */}
      <BrowserRouter>
        <SprintList/>
        <PullRequestList/>
      </BrowserRouter>
    </>
  )
}