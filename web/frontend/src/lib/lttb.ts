// Largest-Triangle-Three-Buckets downsampling — preserves the visual shape
// of the minimap while capping the point count (canvas 2D, no ECharts).
export function lttb(data: Array<[number, number]>, threshold: number): Array<[number, number]> {
  const n = data.length;
  if (threshold >= n || threshold <= 2) return data.slice();
  const sampled: Array<[number, number]> = [data[0]];
  const bucketSize = (n - 2) / (threshold - 2);
  let a = 0;
  for (let i = 0; i < threshold - 2; i++) {
    const rangeStart = Math.floor((i + 1) * bucketSize) + 1;
    const rangeEnd = Math.min(Math.floor((i + 2) * bucketSize) + 1, n);
    const avgRangeStart = Math.floor((i + 0) * bucketSize) + 1;
    const avgRangeEnd = Math.min(Math.floor((i + 1) * bucketSize) + 1, n);
    const avgX = (data[avgRangeStart][0] + data[avgRangeEnd - 1][0]) / 2;
    const avgY = averageY(data, avgRangeStart, avgRangeEnd);
    const [ax, ay] = data[a];
    let maxArea = -1;
    let nextA = rangeStart;
    for (let j = rangeStart; j < rangeEnd; j++) {
      const area = Math.abs((ax - avgX) * (data[j][1] - ay) - (ax - data[j][0]) * (avgY - ay));
      if (area > maxArea) { maxArea = area; nextA = j; }
    }
    sampled.push(data[nextA]);
    a = nextA;
  }
  sampled.push(data[n - 1]);
  return sampled;
}

function averageY(data: Array<[number, number]>, from: number, to: number): number {
  let sum = 0;
  for (let i = from; i < to; i++) sum += data[i][1];
  return sum / (to - from);
}
