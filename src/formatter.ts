import chalk from 'chalk';
import boxen from 'boxen';

export interface FormatOptions {
  plain?: boolean;
  title?: string;
}

export function formatJoke(
  joke: string,
  { plain = false, title = "Today's Dad Joke" }: FormatOptions = {}
): string {
  if (plain) {
    return joke;
  }

  const styledJoke = chalk.italic.cyanBright(`"${joke}"`);
  const styledTitle = chalk.bold.yellow(`🎭 ${title}`);

  return boxen(styledJoke, {
    title: styledTitle,
    titleAlignment: 'left',
    padding: { top: 0, bottom: 0, left: 1, right: 1 },
    margin: { top: 0, bottom: 0, left: 0, right: 0 },
    borderStyle: 'round',
    borderColor: 'yellow'
  });
}
