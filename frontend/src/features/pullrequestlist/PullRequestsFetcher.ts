import { Member, Sprint } from "../sprintlist/SprintRow";
import { PR } from "./PullRequest";

export const FetchPullRequests = async (sprint: Sprint): Promise<PR[]> => {
    const params = {
        startdate : sprint.startDate.toISOString().split('T')[0],
        enddate: sprint.endDate.toISOString().split('T')[0],
    };
    const queryParameters = new URLSearchParams(params);
    sprint.members.map(
        (member: Member) => queryParameters.append('developers', member.name)
    )

    try {
        const res = await fetch(`http://localhost:8080/api/pull_requests?${queryParameters}`);
        const data = await res.json();
        const result_2: PR[] = [];

        for (const pr of data) {
            // PR型に詰め替え
            const newPR: PR = {
                id: pr.Number,
                title: pr.Title,
                branchName: pr.HeadRefName,
                url: pr.URL,
                username: pr.Author.Login,
                iconURL: pr.Author.AvatarURL,
                repository: pr.Repository.Name,
                created: new Date(pr.CreatedAt),
                firstReviewed: new Date(pr.FirstReviewed.Nodes[0].CreatedAt),
                lastApproved: new Date(pr.LastApprovedAt.Nodes[0].CreatedAt),
                merged: new Date(pr.MergedAt),
            };

            // 配列に追加
            result_2.push(newPR);
        }
        return result_2;
    } catch (error) {
        console.error(error);
        return [];
    }
}