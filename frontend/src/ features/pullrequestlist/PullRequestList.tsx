import { PullRequest, PR } from "./PullRequest";

type PullRequestProps = {
    prs: PR[];
}

export const PullRequestList: React.FC<PullRequestProps> = ({prs}) => {
    return (
        <>
            <h1>プルリクエスト一覧</h1>

            <div className="flex overflow-x-auto">
                <table className="flex-none divide-y divide-gray-200 dark:divide-gray-700">
                    <thead>
                        <tr>
                            <th scope="col" className="px-6 py-3 text-start text-xs font-medium text-gray-500 uppercase">ID</th>
                            <th scope="col" className="px-6 py-3 text-start text-xs font-medium text-gray-500 uppercase">Title</th>
                            <th scope="col" className="px-6 py-3 text-start text-xs font-medium text-gray-500 uppercase">URL</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                        {prs?.map((pr: PR) => {
                            return <PullRequest key={pr.id} pr={pr} />;
                        })}
                    </tbody>
                </table>
            </div>
        </>
    )
}