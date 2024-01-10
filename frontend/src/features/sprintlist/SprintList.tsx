import { useEffect, useState } from 'react'
import { Sprint, SprintRow } from './SprintRow'
import { GetSprintList } from './GetConstSprintList';

export const SprintList = () => {
    const [sprintList, setSprintList] = useState<Sprint[]>([])

    useEffect(() => {
        const fetchSprintList = async () => {
            setSprintList(GetSprintList())
        }
        fetchSprintList()
    }, [])

    const sprintListSorted = sprintList.sort((a, b) => b.id - a.id)

    return (
        <table className='divide-y divide-gray-200 dark:divide-gray-700'>
            <thead>
                <tr>
                    <th>No.</th>
                    <th>開始日</th>
                    <th>終了日</th>
                    <th>参加者</th>
                    <th>詳細</th>
                </tr>
            </thead>
            <tbody>
                {sprintListSorted?.map((sprint: Sprint) => {
                    return <SprintRow key={sprint.id} sprint={sprint}/>
                })}
            </tbody>
        </table>
    )
}
