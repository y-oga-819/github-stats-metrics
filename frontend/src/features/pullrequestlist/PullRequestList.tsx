import { useEffect, useState } from "react";
import { PullRequest, PR } from "./PullRequest";
import { Member, Sprint } from "../sprintlist/SprintRow";

type PullRequestListProp = {
    sprint: Sprint
}

export const PullRequestList: React.FC<PullRequestListProp> = ({sprint}) => {
    const [pullRequests, setPullRequests] = useState<PR[]|null>(null);

    useEffect(() => {
        const fetchPullRequests = async () => {
            const params = {
                startdate : sprint.startDate.toISOString().split('T')[0],
                enddate: sprint.endDate.toISOString().split('T')[0],
            };
            const queryParameters = new URLSearchParams(params);
            sprint.members.map(
                (member: Member) => queryParameters.append('developers', member.name)
            )
        
            fetch(`http://localhost:8080/api/pull_requests?${queryParameters}`)
            .then((res) => res.json())
            .then((data) => {
                const result: PR[] = [];
    
                for (const pr of data) {
                    // PR型に詰め替え
                    const newPR: PR = {
                        id: pr.Number,
                        title: pr.Title,
                        url: pr.URL,
                        username: pr.Author.Login,
                        iconURL: pr.Author.AvatarURL,
                        repository: pr.Repository.Name,
                        created: new Date(pr.CreatedAt),
                        firstReviewed: new Date(pr.FirstReviewed.Nodes[0].CreatedAt),
                        lastApproved: new Date(pr.LastApprovedAt.Nodes[0].CreatedAt),
                        merged: new Date(pr.MergedAt),
                    }
    
                    // 配列に追加
                    result.push(newPR)
                }
    
                // Stateにセット
                setPullRequests(result);
            })
            .catch(error => console.error(error));
        }

        fetchPullRequests()
    }, []); // 空配列を依存リストとして渡すと、マウント時に一回だけ実行される

    return (
        <>
            <h1 className="text-left">プルリクエスト一覧</h1>

            <div className="flex overflow-x-auto">
                <table className="flex-none divide-y divide-gray-200 dark:divide-gray-700">
                    <thead>
                        <tr>
                            <th scope="col" className="pl-4 py-2 text-start text-xs font-medium text-gray-500 ">ID</th>
                            <th scope="col" className="pl-4 py-2 text-start text-xs font-medium text-gray-500 ">Author</th>
                            <th scope="col" className="pl-4 py-2 text-start text-xs font-medium text-gray-500 ">初回レビューまで</th>
                            <th scope="col" className="pl-4 py-2 text-start text-xs font-medium text-gray-500 ">初回〜最終Aprvまで</th>
                            <th scope="col" className="pl-4 py-2 text-start text-xs font-medium text-gray-500 ">最終Aprv〜Mergeまで</th>
                            <th scope="col" className="pl-4 py-2 text-start text-xs font-medium text-gray-500 ">Title</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                        {pullRequests?.map((pr: PR) => {
                            return <PullRequest key={pr.id} pr={pr} />;
                        })}
                    </tbody>
                </table>
            </div>
        </>
    )
}