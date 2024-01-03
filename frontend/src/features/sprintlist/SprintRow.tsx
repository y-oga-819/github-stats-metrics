
export type Member = {
    name: String
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
            <td className="whitespace-pre-wrap">
                {sprint.members.map(
                    (member: Member) => {return member.name}
                )
                .join(', ')}
            </td>
        </tr>
    )
}