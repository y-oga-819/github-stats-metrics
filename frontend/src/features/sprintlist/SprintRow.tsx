import { Link } from "react-router-dom"

export type Member = {
    name: string
    iconURL: string
}

export type Sprint = {
    id: number
    startDate: Date
    endDate: Date
    members: Member[]
}

type SprintProp = {
    sprint: Sprint
}

export const SprintRow: React.FC<SprintProp> = ({sprint}) => {
    return (
        <tr>
            <td>{sprint.id}</td>
            <td>{sprint.startDate.toISOString().split('T')[0]}</td>
            <td>{sprint.endDate.toISOString().split('T')[0]}</td>
            <td className="text-left">
                {sprint.members.map(
                    (member: Member) => <img className="inline-block pr-1" src={member.iconURL} width={20} />
                )}
            </td>
            <td><Link to={'/sprints/' + sprint.id}>詳細</Link></td>
        </tr>
    )
}