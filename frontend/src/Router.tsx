import { Route, Routes } from 'react-router-dom';
import { SprintList } from './features/sprintlist/SprintList';
import { Chart } from './features/pullrequestlist/Chart';

export const AppRouter = () => {
    return (
        <Routes>
            <Route path='/' element={<Chart />} />
            <Route path='sprints' >
                <Route index element={<SprintList />} />
            </Route>
        </Routes>
    )
}