#!/usr/bin/env node

import { Command } from 'commander';
import ora from 'ora';
import {
  getDailyJoke,
  getRandomJoke,
  hasCachedJokeForToday
} from './joke-service.js';
import { formatJoke } from './formatter.js';

interface CliOptions {
  refresh?: boolean;
  random?: boolean;
  plain?: boolean;
}

export function createProgram(): Command {
  const program = new Command();

  program
    .name('dadjoke')
    .description('Daily Dad Joke CLI - Welcome your terminal sessions with humor!')
    .version('1.0.0', '-v, --version', 'Display the current version')
    .option('-r, --refresh', 'Force refresh today\'s joke from the API')
    .option('--random', 'Fetch a new random joke immediately without updating daily cache')
    .option('-p, --plain', 'Print plain text joke without box or color formatting')
    .action(async (options: CliOptions) => {
      const forceRefresh = Boolean(options.refresh);
      const isRandom = Boolean(options.random);
      const plain = Boolean(options.plain);

      const willFetchNetwork = isRandom || forceRefresh || !hasCachedJokeForToday();
      const spinner =
        !plain && willFetchNetwork && process.stdout.isTTY
          ? ora({ text: 'Fetching dad joke...', color: 'yellow' }).start()
          : null;

      try {
        let result;
        let title = "Today's Dad Joke";

        if (isRandom) {
          result = await getRandomJoke();
          title = 'Random Dad Joke';
        } else {
          result = await getDailyJoke({ forceRefresh });
          if (forceRefresh) {
            title = "Today's Dad Joke (Refreshed)";
          }
        }

        spinner?.stop();

        if (result && result.joke) {
          console.log(formatJoke(result.joke, { plain, title }));
        }
      } catch (err) {
        spinner?.stop();
        if (process.env.DEBUG) {
          console.error(err);
        }
      }
    });

  program
    .command('init')
    .description('Configure terminal startup hook (PowerShell, Bash, Zsh)')
    .action(async () => {
      console.log('Init command will be implemented in the next task.');
    });

  return program;
}

async function run(): Promise<void> {
  const program = createProgram();
  await program.parseAsync(process.argv);
}

run();
