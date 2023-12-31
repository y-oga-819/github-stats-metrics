import { useEffect, useState } from "react";
import { PullRequest, PR } from "./PullRequest";

export const PullRequestList = () => {
    const [pullRequests, setPullRequests] = useState<PR[]>([]);
  
    useEffect(() => {
        fetch('http://localhost:8080/api/pull_requests')
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
    }, []);

    return (
        <>
            <h1>プルリクエスト一覧</h1>

            <div className="flex overflow-x-auto">
                <table className="flex-none divide-y divide-gray-200 dark:divide-gray-700">
                    <thead>
                        <tr>
                            <th scope="col" className="px-6 py-3 text-start text-xs font-medium text-gray-500 uppercase">ID</th>
                            <th scope="col" className="px-6 py-3 text-start text-xs font-medium text-gray-500 uppercase">Author</th>
                            <th scope="col" className="px-6 py-3 text-start text-xs font-medium text-gray-500 uppercase">Title</th>
                            <th scope="col" className="px-6 py-3 text-start text-xs font-medium text-gray-500 uppercase">初回レビューまでにかかった時間</th>
                            <th scope="col" className="px-6 py-3 text-start text-xs font-medium text-gray-500 uppercase">初回〜最終Aprvまでの時間</th>
                            <th scope="col" className="px-6 py-3 text-start text-xs font-medium text-gray-500 uppercase">最終Aprv〜Mergeまでの時間</th>
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