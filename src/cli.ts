#!/usr/bin/env node

import { Command, InvalidArgumentError } from 'commander';
import chalk from 'chalk';
import boxen from 'boxen';
import ora from 'ora';
import {
  getDailyJoke,
  getRandomJoke,
  hasCachedJokeForToday
} from './joke-service.js';
import { formatJoke } from './formatter.js';
import {
  configureShellHook,
  type SupportedShell
} from './init-service.js';

interface CliOptions {
  refresh?: boolean;
  random?: boolean;
  plain?: boolean;
}

interface InitCommandOptions {
  shell?: SupportedShell;
  profile?: string;
  dryRun?: boolean;
  uninstall?: boolean;
}

function parseShellOption(value: string): SupportedShell {
  const normalized = value.trim().toLowerCase();
  if (normalized === 'powershell' || normalized === 'pwsh' || normalized === 'ps') {
    return 'powershell';
  }
  if (normalized === 'bash') {
    return 'bash';
  }
  if (normalized === 'zsh') {
    return 'zsh';
  }
  throw new InvalidArgumentError('Shell must be one of: powershell, bash, zsh');
}

export function createProgram(): Command {
  const program = new Command();

  program
    .name('dadjoke')
    .description('Daily Dad Joke CLI - Welcome your terminal sessions with humor!')
    .version('1.0.0', '-v, --version', 'Display the current version')
    .option('-r, --refresh', "Force refresh today's joke from the API")
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
    .option('-s, --shell <shell>', 'Target shell (powershell, bash, zsh)', parseShellOption)
    .option('--profile <path>', 'Custom shell profile file path')
    .option('-d, --dry-run', 'Preview the profile hook snippet without writing to disk')
    .option('-u, --uninstall', 'Remove the daily dad joke hook from your shell profile')
    .action((options: InitCommandOptions) => {
      const result = configureShellHook({
        shell: options.shell,
        profilePath: options.profile,
        dryRun: Boolean(options.dryRun),
        uninstall: Boolean(options.uninstall)
      });

      if (result.action === 'dry-run') {
        const body = [
          `${chalk.bold('Shell:')}   ${chalk.cyan(result.shell)}`,
          `${chalk.bold('Profile:')} ${chalk.underline(result.profilePath)}`,
          '',
          chalk.dim(result.snippet)
        ].join('\n');

        console.log(
          boxen(body, {
            title: chalk.bold.yellow('🔍 Shell Hook Preview (Dry Run)'),
            padding: { top: 0, bottom: 0, left: 1, right: 1 },
            borderStyle: 'round',
            borderColor: 'cyan'
          })
        );
        return;
      }

      if (result.action === 'uninstalled') {
        console.log(
          chalk.green(`✔ Removed daily dad joke hook from ${chalk.underline(result.profilePath)}`)
        );
        return;
      }

      if (result.action === 'not-installed') {
        console.log(
          chalk.yellow(`ℹ No daily dad joke hook found in ${chalk.underline(result.profilePath)}`)
        );
        return;
      }

      const statusLabel =
        result.action === 'already-installed'
          ? chalk.green('✔ Startup hook is already up to date!')
          : result.action === 'updated'
            ? chalk.green('✔ Updated startup hook in your shell profile!')
            : chalk.green('✔ Installed startup hook in your shell profile!');

      const summary = [
        statusLabel,
        `${chalk.bold('Shell:')}   ${chalk.cyan(result.shell)}`,
        `${chalk.bold('Profile:')} ${chalk.underline(result.profilePath)}`,
        '',
        chalk.dim('Open a new terminal session (or reload your profile) to see today\'s joke!')
      ].join('\n');

      console.log(
        boxen(summary, {
          title: chalk.bold.yellow('🎭 Daily Dad Joke Init'),
          padding: { top: 0, bottom: 0, left: 1, right: 1 },
          borderStyle: 'round',
          borderColor: 'green'
        })
      );
    });

  return program;
}

async function run(): Promise<void> {
  const program = createProgram();
  await program.parseAsync(process.argv);
}

const isDirectExecution =
  process.argv[1] &&
  (process.argv[1].endsWith('cli.ts') ||
    process.argv[1].endsWith('cli.js') ||
    process.argv[1].endsWith('dadjoke'));

if (isDirectExecution) {
  run();
}
