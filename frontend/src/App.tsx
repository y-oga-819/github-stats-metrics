import './App.css'
import { BrowserRouter } from 'react-router-dom';
import { SprintList } from './features/sprintlist/SprintList';

export const App = () => {
  return (
    <>
      <BrowserRouter>
        <SprintList/>
      </BrowserRouter>
    </>
  )
}