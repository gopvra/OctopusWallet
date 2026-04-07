const colors: Record<string, string> = {
  pending: 'bg-yellow-100 text-yellow-800',
  confirming: 'bg-blue-100 text-blue-800',
  completed: 'bg-green-100 text-green-800',
  expired: 'bg-gray-100 text-gray-800',
  failed: 'bg-red-100 text-red-800',
  processing: 'bg-blue-100 text-blue-800',
  rejected: 'bg-red-100 text-red-800',
  approved: 'bg-green-100 text-green-800',
  pending_approval: 'bg-orange-100 text-orange-800',
};

export function StatusBadge({ status }: { status: string }) {
  const color = colors[status] || 'bg-gray-100 text-gray-800';
  return (
    <span className={`px-2 py-1 rounded-full text-xs font-medium ${color}`}>
      {status.replace('_', ' ')}
    </span>
  );
}
