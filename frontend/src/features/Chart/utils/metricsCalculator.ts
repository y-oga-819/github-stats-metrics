import { PullRequest } from '../../pullrequestlist/types';
import { Sprint } from '../../sprintlist/SprintRow';
import { Metrics } from '../Chart';

/**
 * メトリクス計算ユーティリティ
 * 複数のチャートで共通するメトリクス計算ロジックを集約
 */
export class MetricsCalculator {
  /**
   * 初回レビューまでの平均時間を計算（秒）
   */
  static calculateUntilFirstReviewed(pullRequests: PullRequest[]): number {
    if (pullRequests.length === 0) return 0;
    
    const validPRs = pullRequests.filter(pr => pr.firstReviewed !== null);
    if (validPRs.length === 0) return 0;
    
    const totalTime = validPRs
      .map(pr => (pr.firstReviewed!.getTime() - pr.created.getTime()) / 1000)
      .reduce((sum, time) => sum + time, 0);
    
    return totalTime / validPRs.length;
  }

  /**
   * 最終承認までの平均時間を計算（秒）
   */
  static calculateUntilLastApproved(pullRequests: PullRequest[]): number {
    if (pullRequests.length === 0) return 0;
    
    const validPRs = pullRequests.filter(pr => pr.lastApproved !== null && pr.firstReviewed !== null);
    if (validPRs.length === 0) return 0;
    
    const totalTime = validPRs
      .map(pr => (pr.lastApproved!.getTime() - pr.firstReviewed!.getTime()) / 1000)
      .reduce((sum, time) => sum + time, 0);
    
    return totalTime / validPRs.length;
  }

  /**
   * マージまでの平均時間を計算（秒）
   */
  static calculateUntilMerged(pullRequests: PullRequest[]): number {
    if (pullRequests.length === 0) return 0;
    
    const validPRs = pullRequests.filter(pr => pr.merged !== null && pr.lastApproved !== null);
    if (validPRs.length === 0) return 0;
    
    const totalTime = validPRs
      .map(pr => (pr.merged!.getTime() - pr.lastApproved!.getTime()) / 1000)
      .reduce((sum, time) => sum + time, 0);
    
    return totalTime / validPRs.length;
  }

  /**
   * Dev/Day/Developer指標を計算
   */
  static calculateDevDayDeveloper(pullRequestCount: number, memberCount: number): number {
    if (memberCount === 0) return 0;
    return pullRequestCount / memberCount / 5; // 5日間での計算
  }

  /**
   * スプリント用メトリクスを一括計算
   */
  static calculateSprintMetrics(
    sprintId: number,
    pullRequests: PullRequest[],
    sprint: Sprint
  ) {
    const prCount = pullRequests.length;
    
    return {
      prCount: { sprintId, score: prCount } as Metrics,
      devDayDeveloper: { 
        sprintId, 
        score: this.calculateDevDayDeveloper(prCount, sprint.members.length) 
      } as Metrics,
      untilFirstReviewed: { 
        sprintId, 
        score: this.calculateUntilFirstReviewed(pullRequests) 
      } as Metrics,
      untilLastApproved: { 
        sprintId, 
        score: this.calculateUntilLastApproved(pullRequests) 
      } as Metrics,
      untilMerged: { 
        sprintId, 
        score: this.calculateUntilMerged(pullRequests) 
      } as Metrics,
    };
  }
}