import { useEffect, useState } from 'react'
import { Sprint, SprintRow } from './SprintRow'

export const SprintList = () => {
    const result: Sprint[] = [];

    const [sprintList, setSprintList] = useState<Sprint[]>([])

    useEffect(() => {
        result.push({
            id: 1,
            startDate: new Date('2024-01-01'),
            endDate: new Date('2024-01-08'),
            members: [
                {name: 'oga'}, {name: 'shiiyan'}
            ]
        },
        {
            id: 2,
            startDate: new Date('2024-01-09'),
            endDate: new Date('2024-01-16'),
            members: [
                {name: 'oga'}, {name: 'shiiyan'}
            ]
        })
    
        setSprintList(result)    
    })

    return (
        <table>
            <thead>
                <tr>
                    <th>No.</th>
                    <th>開始日</th>
                    <th>終了日</th>
                    <th>参加者</th>
                </tr>
            </thead>
            <tbody>
                {sprintList?.map((sprint: Sprint) => {
                    return <SprintRow key={sprint.id} sprint={sprint}/>
                })}
            </tbody>
        </table>
    )
}