import { useLocation } from "react-router-dom"
import { PullRequestList } from "../pullrequestlist/PullRequestList"
import { Sprint } from "../sprintlist/SprintRow"

interface State {
    sprint: Sprint
}

export const SprintDetail = () => {
    const location = useLocation();
    const { sprint } = location.state as State;

    return (
        <>
            <h2>{'Sprint ' + sprint.id}</h2>
            <PullRequestList sprint={sprint}/>
        </>
    )
}