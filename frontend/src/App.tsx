import { NavLink } from 'react-router-dom';
import './App.css'
import { AppRouter } from './Router';

export const App = () => {
  return (
    <>
      <h1 className='pl-24 font-bold'>GitHub Metrics Tracker</h1>
      <aside className='fixed top-0 left-0 h-screen'>
        <div className='h-full px-8 py-4 bg-gray-100'>
          <h2 className='py-1 font-bold'>Menu</h2>
          <ul>
            <li><NavLink to='/'>Chart</NavLink></li>
            <li><NavLink to='/sprints'>Sprints</NavLink></li>
          </ul>
        </div>
      </aside>
      <div className='pl-24'>
        <AppRouter />
      </div>
    </>
  )
}