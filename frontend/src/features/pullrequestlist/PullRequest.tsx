import {Link} from 'react-router-dom'

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
    return (
      <tr className="divide-x divide-gray-200">
        <td className="whitespace-nowrap py-4 pl-4 pr-4 text-sm text-gray-500 sm:pl-0">{pr.repository}:{pr.id}</td>
        <td className="whitespace-nowrap p-4 text-sm text-gray-500"><img className='object-contain' width="20" src={pr.iconURL}/></td>
        <td className="whitespace-nowrap text-right p-4 text-sm font-semibold text-gray-600">{format(pr.created, pr.firstReviewed)}</td>
        <td className="whitespace-nowrap text-right p-4 text-sm font-semibold text-gray-600">{format(pr.firstReviewed, pr.lastApproved)}</td>
        <td className="whitespace-nowrap text-right p-4 text-sm font-semibold text-gray-600">{format(pr.lastApproved, pr.merged)}</td>
        <td className="whitespace-nowrap text-left p-4 text-sm text-gray-500"><Link to={pr.url}>{pr.title}</Link></td>
      </tr>
    )
}
