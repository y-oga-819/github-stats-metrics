import {Link} from 'react-router-dom'

type PullRequestProps = {
    pr: PR
}

export type PR = {
    id: number
    title: string
    url: string
    username: string
    iconURL: string
    repository: string
    created: Date
    firstReviewed: Date
    lastApproved: Date
    merged: Date
}

export const PullRequest: React.FC<PullRequestProps> = ({pr}) => {
    return (
        <tr className="hover:bg-gray-100 dark:hover:bg-gray-700">
            <td className="text-xs text-left">{pr.repository}:{pr.id}</td>
            <td><img className='object-contain' width="20" src={pr.iconURL}/></td>
            <td className='text-left text-xs'><Link to={pr.url}>{pr.title}</Link></td>
            <td>{(pr.firstReviewed.getTime() - pr.created.getTime()).toString()}</td>
            <td>{(pr.lastApproved.getTime() - pr.firstReviewed.getTime()).toString()}</td>
            <td>{(pr.merged.getTime() - pr.lastApproved.getTime()).toString()}</td>
            <td></td>
        </tr>
    )
}