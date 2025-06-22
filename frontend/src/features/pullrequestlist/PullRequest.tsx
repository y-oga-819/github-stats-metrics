import {Link} from 'react-router-dom'

/**
 * 日付差分を安全にフォーマット（null安全）
 */
function format(before: Date | null, after: Date | null): string {
  if (!before || !after) {
    return '未設定';
  }

  // 日付の妥当性チェック
  if (isNaN(before.getTime()) || isNaN(after.getTime())) {
    return '無効な日付';
  }

  // 負の時間差の場合
  if (after.getTime() < before.getTime()) {
    return '日付エラー';
  }

  let sec = Math.floor((after.getTime() - before.getTime()) / 1000);

  if (sec < 0) {
    return '計算エラー';
  }

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

/**
 * 安全な画像表示（エラー時の代替表示）
 */
function SafeAvatar({ src, alt, className }: { src: string; alt: string; className?: string }) {
  const handleImageError = (e: React.SyntheticEvent<HTMLImageElement>) => {
    e.currentTarget.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAiIGhlaWdodD0iMjAiIHZpZXdCb3g9IjAgMCAyMCAyMCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPGNpcmNsZSBjeD0iMTAiIGN5PSIxMCIgcj0iMTAiIGZpbGw9IiNlNWU3ZWIiLz4KPHN2ZyB3aWR0aD0iMTIiIGhlaWdodD0iMTIiIHZpZXdCb3g9IjAgMCAxMiAxMiIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIiB4PSI0IiB5PSI0Ij4KPHN2ZyB3aWR0aD0iMTIiIGhlaWdodD0iMTIiIHZpZXdCb3g9IjAgMCAxMiAxMiIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPGNpcmNsZSBjeD0iNiIgY3k9IjQiIHI9IjIiIGZpbGw9IiM5Y2ExYTYiLz4KPHBhdGggZD0iTTIgMTBDMiA4IDQgNyA2IDcgOCA3IDEwIDggMTAgMTAiIHN0cm9rZT0iIzljYTFhNiIgc3Ryb2tlLWxpbmVjYXA9InJvdW5kIi8+Cjwvc3ZnPgo8L3N2Zz4KPC9zdmc+'; // デフォルトのユーザーアバター
    e.currentTarget.alt = 'Default Avatar';
  };

  return (
    <img 
      className={className} 
      width="20" 
      src={src} 
      alt={alt}
      onError={handleImageError}
    />
  );
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
    firstReviewed: Date | null
    lastApproved: Date | null
    merged: Date | null
}

export const PullRequest: React.FC<PullRequestProps> = ({pr}) => {
    // 必須フィールドの検証
    if (!pr || typeof pr.id !== 'number' || !pr.title || !pr.repository) {
        return (
            <tr className="divide-x divide-gray-200 bg-red-50">
                <td colSpan={6} className="whitespace-nowrap py-4 pl-4 pr-4 text-sm text-red-500 sm:pl-0">
                    無効なPull Requestデータ
                </td>
            </tr>
        );
    }

    return (
      <tr className="divide-x divide-gray-200">
        <td className="whitespace-nowrap py-4 pl-4 pr-4 text-sm text-gray-500 sm:pl-0">
            {pr.repository}:{pr.id}
        </td>
        <td className="whitespace-nowrap p-4 text-sm text-gray-500">
            <SafeAvatar 
                src={pr.iconURL} 
                alt={`${pr.username} avatar`}
                className="object-contain"
            />
        </td>
        <td className="whitespace-nowrap text-right p-4 text-sm font-semibold text-gray-600">
            {format(pr.created, pr.firstReviewed)}
        </td>
        <td className="whitespace-nowrap text-right p-4 text-sm font-semibold text-gray-600">
            {format(pr.firstReviewed, pr.lastApproved)}
        </td>
        <td className="whitespace-nowrap text-right p-4 text-sm font-semibold text-gray-600">
            {format(pr.lastApproved, pr.merged)}
        </td>
        <td className="whitespace-nowrap text-left p-4 text-sm text-gray-500">
            {pr.url ? (
                <Link to={pr.url} className="text-blue-600 hover:text-blue-800">
                    {pr.title}
                </Link>
            ) : (
                <span className="text-gray-400">{pr.title}</span>
            )}
        </td>
      </tr>
    )
}
