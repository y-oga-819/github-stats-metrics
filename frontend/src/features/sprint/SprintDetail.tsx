import { useParams } from "react-router-dom"
import { PullRequestList } from "../pullrequestlist/PullRequestList"
type Params = {
    sprintId?: string
}

export const SprintDetail = () => {
    const params: Params = useParams<Params>()
    return (
        <>
            <h2>{'Sprint ' + params?.sprintId}</h2>
            <PullRequestList/>
        </>
    )
}