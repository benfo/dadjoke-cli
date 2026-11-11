export const FALLBACK_JOKES: readonly string[] = [
  "Why do fathers take an extra pair of socks golfing? In case they get a hole in one.",
  "I told my doctor that I broke my arm in two places. He told me to stop going to those places.",
  "What do you call a fake noodle? An impasta.",
  "How does a penguin build its house? Igloos it together.",
  "Why don't skeletons fight each other? They don't have the guts.",
  "What do you call cheese that isn't yours? Nacho cheese.",
  "Why did the scarecrow win an award? Because he was outstanding in his field.",
  "I'm reading a book on anti-gravity. It's impossible to put down.",
  "Why do melons have weddings? Because they cantaloupe.",
  "What did the ocean say to the shore? Nothing, it just waved."
];

export function getFallbackJoke(seed: string): string {
  let hash = 0;
  for (let i = 0; i < seed.length; i++) {
    hash = (hash * 31 + seed.charCodeAt(i)) >>> 0;
  }
  const index = hash % FALLBACK_JOKES.length;
  return FALLBACK_JOKES[index];
}
