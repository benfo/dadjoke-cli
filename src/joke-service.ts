import fs from 'fs-extra';
import path from 'node:path';
import os from 'node:os';
import { getFallbackJoke } from './fallbacks.js';

export const JOKE_API_URL = 'https://icanhazdadjoke.com/';

export interface CachedJokeRecord {
  date: string;
  joke: string;
  id?: string | null;
  cachedAt?: string;
}

export type JokeSource = 'cache' | 'network' | 'stale-cache' | 'offline-fallback';

export interface JokeResult {
  joke: string;
  source: JokeSource;
  date?: string;
}

export interface FetchOptions {
  timeoutMs?: number;
  fetchFn?: typeof globalThis.fetch;
}

export interface DailyJokeOptions extends FetchOptions {
  forceRefresh?: boolean;
  cachePath?: string;
  todayKey?: string;
}

export function getTodayKey(date: Date = new Date()): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

export function getDefaultCachePath(): string {
  return path.join(os.homedir(), '.dadjoke-cache.json');
}

export function readCache(cachePath: string = getDefaultCachePath()): CachedJokeRecord | null {
  try {
    if (!fs.pathExistsSync(cachePath)) {
      return null;
    }
    const data = fs.readJsonSync(cachePath) as Partial<CachedJokeRecord>;
    if (data && typeof data.joke === 'string' && typeof data.date === 'string') {
      return {
        date: data.date,
        joke: data.joke,
        id: data.id ?? null,
        cachedAt: data.cachedAt
      };
    }
  } catch {
    // Ignore cache corruption/read errors
  }
  return null;
}

export function writeCache(
  cacheData: CachedJokeRecord,
  cachePath: string = getDefaultCachePath()
): void {
  try {
    fs.outputJsonSync(cachePath, cacheData, { spaces: 2 });
  } catch {
    // Non-fatal if cache cannot be written
  }
}

export async function fetchRemoteDadJoke({
  timeoutMs = 3000,
  fetchFn = globalThis.fetch
}: FetchOptions = {}): Promise<{ id: string | null; joke: string } | null> {
  try {
    const response = await fetchFn(JOKE_API_URL, {
      headers: {
        Accept: 'application/json',
        'User-Agent': 'DailyDadJokeCLI (https://github.com/personal/wk-testing-dadjokes)'
      },
      signal: AbortSignal.timeout(timeoutMs)
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }

    const json = (await response.json()) as { id?: string; joke?: string };
    if (json && typeof json.joke === 'string') {
      return {
        id: json.id ?? null,
        joke: json.joke.trim()
      };
    }
    throw new Error('Invalid JSON payload');
  } catch {
    return null;
  }
}

export function hasCachedJokeForToday(
  cachePath: string = getDefaultCachePath(),
  todayKey: string = getTodayKey()
): boolean {
  const cached = readCache(cachePath);
  return Boolean(cached && cached.date === todayKey && cached.joke);
}

export async function getDailyJoke({
  forceRefresh = false,
  cachePath = getDefaultCachePath(),
  todayKey = getTodayKey(),
  fetchFn = globalThis.fetch,
  timeoutMs = 3000
}: DailyJokeOptions = {}): Promise<JokeResult> {
  if (!forceRefresh) {
    const cached = readCache(cachePath);
    if (cached && cached.date === todayKey && cached.joke) {
      return {
        joke: cached.joke,
        source: 'cache',
        date: cached.date
      };
    }
  }

  const remote = await fetchRemoteDadJoke({ timeoutMs, fetchFn });
  if (remote && remote.joke) {
    const cacheRecord: CachedJokeRecord = {
      date: todayKey,
      joke: remote.joke,
      id: remote.id,
      cachedAt: new Date().toISOString()
    };
    writeCache(cacheRecord, cachePath);
    return {
      joke: remote.joke,
      source: 'network',
      date: todayKey
    };
  }

  const existing = readCache(cachePath);
  if (existing && existing.joke) {
    return {
      joke: existing.joke,
      source: 'stale-cache',
      date: existing.date
    };
  }

  return {
    joke: getFallbackJoke(todayKey),
    source: 'offline-fallback',
    date: todayKey
  };
}

export async function getRandomJoke({
  fetchFn = globalThis.fetch,
  timeoutMs = 3000
}: FetchOptions = {}): Promise<JokeResult> {
  const remote = await fetchRemoteDadJoke({ timeoutMs, fetchFn });
  if (remote && remote.joke) {
    return {
      joke: remote.joke,
      source: 'network'
    };
  }
  const randomKey = Math.random().toString(36).slice(2);
  return {
    joke: getFallbackJoke(randomKey),
    source: 'offline-fallback'
  };
}
