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
            <td className="px-4 py-1">{sprint.startDate.toISOString().split('T')[0]}</td>
            <td className="px-4 py-1">{sprint.endDate.toISOString().split('T')[0]}</td>
            <td className="text-left px-4">
                {sprint.members.map(
                    (member: Member) => <img key={member.name} className="inline-block pr-1" src={member.iconURL} width={20} />
                )}
            </td>
            <td className="text-blue-800"><Link to={sprint.id.toString()} state={{sprint: sprint}}>詳細</Link></td>
        </tr>
    )
}
