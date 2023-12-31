import './App.css'
import { PullRequestList } from './ features/pullrequestlist/PullRequestList';
import { BrowserRouter } from 'react-router-dom';

function App() {
  return (
    <>
      {/* <Chart prs={pullRequests}/> */}
      <BrowserRouter>
        <PullRequestList/>
      </BrowserRouter>
    </>
  )
}

export default App
