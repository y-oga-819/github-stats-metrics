import {Link} from 'react-router-dom'

type PullRequestProps = {
    pr: PR
}

export type PR = {
    id: number
    title: string
    branchName: string
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
    function format(before: Date, after: Date): string {
        let sec = (after.getTime() - before.getTime()) / 1000;

        let day = Math.floor(sec / 86400);
        let hour = Math.floor(sec % 86400 / 3600);
        let min = Math.floor(sec % 3600 / 60);
        let rem = sec % 60;

        var str = "";
        if (day > 0) str += `${day}日 `;
        if (hour > 0) str += `${hour}時間 `;
        if (min > 0) str += `${min}分 `;
        str += `${rem}秒`;

        return str;
    }

    return (
        <tr className="hover:bg-gray-100 dark:hover:bg-gray-700">
            <td className="text-xs text-left">{pr.repository}:{pr.id}</td>
            <td className='text-center'><img className='object-contain' width="20" src={pr.iconURL}/></td>
            <td className="text-right text-xs">{format(pr.created, pr.firstReviewed)}</td>
            <td className="text-right text-xs">{format(pr.firstReviewed, pr.lastApproved)}</td>
            <td className="text-right text-xs">{format(pr.lastApproved, pr.merged)}</td>
            <td className='text-left text-xs pl-8'><Link to={pr.url}>{pr.title}</Link></td>
            <td></td>
        </tr>
    )
}