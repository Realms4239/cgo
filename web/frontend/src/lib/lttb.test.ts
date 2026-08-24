import { describe, expect, it } from 'vitest';
import { lttb } from './lttb';

describe('lttb', () => {
  it('returns the same array when under the threshold', () => {
    const data: Array<[number, number]> = [[1, 1], [2, 2], [3, 3]];
    expect(lttb(data, 5)).toEqual(data);
  });

  it('keeps the first and last points and caps the length', () => {
    const data: Array<[number, number]> = Array.from({ length: 100 }, (_, i) => [i, Math.sin(i / 5)]);
    const out = lttb(data, 10);
    expect(out[0]).toEqual(data[0]);
    expect(out[out.length - 1]).toEqual(data[99]);
    expect(out.length).toBe(10);
  });

  it('preserves a spike', () => {
    const data: Array<[number, number]> = [
      [0, 0], [1, 0], [2, 0], [3, 100], [4, 0], [5, 0], [6, 0],
    ];
    const out = lttb(data, 5);
    expect(out.some(([, y]) => y === 100)).toBe(true);
  });
});
