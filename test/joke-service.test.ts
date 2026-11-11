import { describe, it, expect, vi } from 'vitest';
import fs from 'fs-extra';
import path from 'node:path';
import os from 'node:os';
import {
  getTodayKey,
  getDailyJoke,
  getRandomJoke,
  readCache,
  writeCache,
  hasCachedJokeForToday
} from '../src/joke-service.js';
import { getFallbackJoke, FALLBACK_JOKES } from '../src/fallbacks.js';
import { formatJoke } from '../src/formatter.js';

describe('daily-dadjoke service', () => {
  it('getTodayKey returns YYYY-MM-DD format', () => {
    const d = new Date(2026, 10, 11); // Nov 11, 2026
    expect(getTodayKey(d)).toBe('2026-11-11');
  });

  it('getFallbackJoke returns valid joke from fallback array', () => {
    const joke = getFallbackJoke('2026-11-11');
    expect(typeof joke).toBe('string');
    expect(joke.length).toBeGreaterThan(0);
    expect(FALLBACK_JOKES).toContain(joke);
  });

  it('readCache and writeCache handle file round-trip via fs-extra', () => {
    const tmpCache = path.join(os.tmpdir(), `test-cache-${Date.now()}.json`);
    try {
      expect(readCache(tmpCache)).toBeNull();
      const data = { date: '2026-11-11', joke: 'Funny dad joke', id: 'abc', views: 1, maxViews: 3 };
      writeCache(data, tmpCache);
      const read = readCache(tmpCache);
      expect(read).toEqual(data);
      expect(hasCachedJokeForToday(tmpCache, '2026-11-11')).toBe(true);
      expect(hasCachedJokeForToday(tmpCache, '2026-11-12')).toBe(false);
    } finally {
      fs.removeSync(tmpCache);
    }
  });

  it('getDailyJoke fetches from API and writes cache, then serves from cache', async () => {
    const tmpCache = path.join(os.tmpdir(), `test-cache-fetch-${Date.now()}.json`);
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ id: '123', joke: 'Mocked API Joke' })
    });

    try {
      const result1 = await getDailyJoke({
        cachePath: tmpCache,
        todayKey: '2026-11-11',
        fetchFn: mockFetch as unknown as typeof globalThis.fetch
      });
      expect(result1.joke).toBe('Mocked API Joke');
      expect(result1.source).toBe('network');
      expect(result1.views).toBe(1);
      expect(result1.hidden).toBe(false);
      expect(mockFetch).toHaveBeenCalledTimes(1);

      const result2 = await getDailyJoke({
        cachePath: tmpCache,
        todayKey: '2026-11-11',
        fetchFn: mockFetch as unknown as typeof globalThis.fetch
      });
      expect(result2.joke).toBe('Mocked API Joke');
      expect(result2.source).toBe('cache');
      expect(result2.views).toBe(2);
      expect(result2.hidden).toBe(false);
      expect(mockFetch).toHaveBeenCalledTimes(1);

      const result3 = await getDailyJoke({
        cachePath: tmpCache,
        todayKey: '2026-11-11',
        forceRefresh: true,
        fetchFn: mockFetch as unknown as typeof globalThis.fetch
      });
      expect(result3.joke).toBe('Mocked API Joke');
      expect(result3.source).toBe('network');
      expect(result3.views).toBe(1);
      expect(result3.hidden).toBe(false);
      expect(mockFetch).toHaveBeenCalledTimes(2);
    } finally {
      fs.removeSync(tmpCache);
    }
  });

  it('hides daily joke after max daily views is exceeded unless forceShow is true', async () => {
    const tmpCache = path.join(os.tmpdir(), `test-cache-views-${Date.now()}.json`);
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ id: '123', joke: 'Limited Joke' })
    });

    try {
      // Session 1 (maxViews = 2)
      const s1 = await getDailyJoke({
        cachePath: tmpCache,
        todayKey: '2026-11-11',
        maxViews: 2,
        fetchFn: mockFetch as unknown as typeof globalThis.fetch
      });
      expect(s1.views).toBe(1);
      expect(s1.hidden).toBe(false);

      // Session 2
      const s2 = await getDailyJoke({
        cachePath: tmpCache,
        todayKey: '2026-11-11',
        fetchFn: mockFetch as unknown as typeof globalThis.fetch
      });
      expect(s2.views).toBe(2);
      expect(s2.hidden).toBe(false);

      // Session 3 -> exceeds maxViews (2) -> hidden!
      const s3 = await getDailyJoke({
        cachePath: tmpCache,
        todayKey: '2026-11-11',
        fetchFn: mockFetch as unknown as typeof globalThis.fetch
      });
      expect(s3.views).toBe(3);
      expect(s3.hidden).toBe(true);

      // Session 4 with forceShow -> shown!
      const s4 = await getDailyJoke({
        cachePath: tmpCache,
        todayKey: '2026-11-11',
        forceShow: true,
        fetchFn: mockFetch as unknown as typeof globalThis.fetch
      });
      expect(s4.views).toBe(4);
      expect(s4.hidden).toBe(false);

      // Next day -> resets views to 1 and shows again!
      const nextDay = await getDailyJoke({
        cachePath: tmpCache,
        todayKey: '2026-11-12',
        fetchFn: mockFetch as unknown as typeof globalThis.fetch
      });
      expect(nextDay.views).toBe(1);
      expect(nextDay.hidden).toBe(false);
    } finally {
      fs.removeSync(tmpCache);
    }
  });

  it('getDailyJoke gracefully falls back when fetch fails', async () => {
    const tmpCache = path.join(os.tmpdir(), `test-cache-fail-${Date.now()}.json`);
    const failingFetch = vi.fn().mockRejectedValue(new Error('Network offline'));

    try {
      const result = await getDailyJoke({
        cachePath: tmpCache,
        todayKey: '2026-11-11',
        fetchFn: failingFetch as unknown as typeof globalThis.fetch
      });
      expect(result.joke).toBeTruthy();
      expect(result.source).toBe('offline-fallback');
    } finally {
      fs.removeSync(tmpCache);
    }
  });

  it('getRandomJoke returns joke from network or fallback', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ id: '999', joke: 'Random Joke' })
    });

    const res = await getRandomJoke({
      fetchFn: mockFetch as unknown as typeof globalThis.fetch
    });
    expect(res.joke).toBe('Random Joke');
    expect(res.source).toBe('network');
  });

  it('formatJoke formats text properly in plain and boxed modes', () => {
    const raw = 'Why did the chicken cross the road?';
    const plain = formatJoke(raw, { plain: true });
    expect(plain).toBe(raw);

    const styled = formatJoke(raw, { plain: false });
    expect(styled).toContain("Today's Dad Joke");
    expect(styled).toContain(raw);
    expect(styled).toContain('╭'); // boxen round border
  });
});
