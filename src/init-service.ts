import fs from 'fs-extra';
import path from 'node:path';
import os from 'node:os';
import { fileURLToPath } from 'node:url';

export type SupportedShell = 'powershell' | 'bash' | 'zsh';

export const HOOK_START_MARKER = '# >>> daily-dadjoke initialize >>>';
export const HOOK_END_MARKER = '# <<< daily-dadjoke initialize <<<';

export interface InitOptions {
  shell?: SupportedShell;
  profilePath?: string;
  cliScriptPath?: string;
  dryRun?: boolean;
  uninstall?: boolean;
  homeDir?: string;
  platform?: NodeJS.Platform;
  env?: NodeJS.ProcessEnv;
}

export interface InitResult {
  shell: SupportedShell;
  profilePath: string;
  snippet: string;
  action: 'installed' | 'updated' | 'already-installed' | 'uninstalled' | 'not-installed' | 'dry-run';
}

export function getDefaultCliScriptPath(): string {
  const currentFile = fileURLToPath(import.meta.url);
  const dir = path.dirname(currentFile);
  if (path.basename(dir) === 'src') {
    return path.resolve(dir, '..', 'dist', 'cli.js');
  }
  return path.resolve(dir, 'cli.js');
}

export function detectShell(
  platform: NodeJS.Platform = process.platform,
  env: NodeJS.ProcessEnv = process.env
): SupportedShell {
  const shellEnv = (env.SHELL || '').toLowerCase();
  if (shellEnv.includes('zsh')) {
    return 'zsh';
  }
  if (shellEnv.includes('bash')) {
    return 'bash';
  }
  if (shellEnv.includes('pwsh') || shellEnv.includes('powershell')) {
    return 'powershell';
  }
  if (platform === 'win32') {
    return 'powershell';
  }
  return 'bash';
}

export function resolveProfilePath(
  shell: SupportedShell,
  {
    homeDir = os.homedir(),
    platform = process.platform,
    env = process.env
  }: {
    homeDir?: string;
    platform?: NodeJS.Platform;
    env?: NodeJS.ProcessEnv;
  } = {}
): string {
  if (shell === 'powershell') {
    if (env.PROFILE && env.PROFILE.trim().length > 0) {
      return env.PROFILE.trim();
    }
    if (platform === 'win32') {
      const pwsh7Dir = path.join(homeDir, 'Documents', 'PowerShell');
      const winPsDir = path.join(homeDir, 'Documents', 'WindowsPowerShell');
      if (!fs.pathExistsSync(pwsh7Dir) && fs.pathExistsSync(winPsDir)) {
        return path.join(winPsDir, 'Microsoft.PowerShell_profile.ps1');
      }
      return path.join(pwsh7Dir, 'Microsoft.PowerShell_profile.ps1');
    }
    return path.join(homeDir, '.config', 'powershell', 'Microsoft.PowerShell_profile.ps1');
  }

  if (shell === 'zsh') {
    return path.join(homeDir, '.zshrc');
  }

  // bash
  const bashrc = path.join(homeDir, '.bashrc');
  const bashProfile = path.join(homeDir, '.bash_profile');
  if (platform === 'darwin' && !fs.pathExistsSync(bashrc) && fs.pathExistsSync(bashProfile)) {
    return bashProfile;
  }
  return bashrc;
}

export function buildHookSnippet(
  shell: SupportedShell,
  cliScriptPath: string = getDefaultCliScriptPath()
): string {
  const normalizedPath = cliScriptPath.replace(/\\/g, '/');

  if (shell === 'powershell') {
    const escapedWinPath = cliScriptPath.replace(/"/g, '`"');
    return [
      HOOK_START_MARKER,
      `if (Test-Path "${escapedWinPath}") { node "${escapedWinPath}" } elseif (Get-Command dadjoke -ErrorAction SilentlyContinue) { dadjoke }`,
      HOOK_END_MARKER
    ].join('\n');
  }

  return [
    HOOK_START_MARKER,
    `if [ -f "${normalizedPath}" ]; then node "${normalizedPath}"; elif command -v dadjoke >/dev/null 2>&1; then dadjoke; fi`,
    HOOK_END_MARKER
  ].join('\n');
}

function escapeRegExp(str: string): string {
  return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

export function applyHookToContent(
  existingContent: string,
  snippet: string
): { content: string; status: 'installed' | 'updated' | 'already-installed' } {
  const blockRegex = new RegExp(
    `${escapeRegExp(HOOK_START_MARKER)}[\\s\\S]*?${escapeRegExp(HOOK_END_MARKER)}`,
    'm'
  );

  if (blockRegex.test(existingContent)) {
    const match = existingContent.match(blockRegex);
    if (match && match[0].replace(/\r\n/g, '\n') === snippet.replace(/\r\n/g, '\n')) {
      return { content: existingContent, status: 'already-installed' };
    }
    const updated = existingContent.replace(blockRegex, snippet);
    return { content: updated, status: 'updated' };
  }

  const trimmed = existingContent.trimEnd();
  const prefix = trimmed.length > 0 ? `${trimmed}\n\n` : '';
  return {
    content: `${prefix}${snippet}\n`,
    status: 'installed'
  };
}

export function removeHookFromContent(
  existingContent: string
): { content: string; removed: boolean } {
  const blockRegex = new RegExp(
    `\\n?${escapeRegExp(HOOK_START_MARKER)}[\\s\\S]*?${escapeRegExp(HOOK_END_MARKER)}\\n?`,
    'm'
  );

  if (!blockRegex.test(existingContent)) {
    return { content: existingContent, removed: false };
  }

  const updated = existingContent.replace(blockRegex, '\n').replace(/\n{3,}/g, '\n\n').trimStart();
  return { content: updated, removed: true };
}

export function configureShellHook(options: InitOptions = {}): InitResult {
  const shell = options.shell ?? detectShell(options.platform, options.env);
  const profilePath =
    options.profilePath ??
    resolveProfilePath(shell, {
      homeDir: options.homeDir,
      platform: options.platform,
      env: options.env
    });
  const cliScriptPath = options.cliScriptPath ?? getDefaultCliScriptPath();
  const snippet = buildHookSnippet(shell, cliScriptPath);

  if (options.dryRun) {
    return {
      shell,
      profilePath,
      snippet,
      action: 'dry-run'
    };
  }

  const existing = fs.pathExistsSync(profilePath)
    ? fs.readFileSync(profilePath, 'utf8')
    : '';

  if (options.uninstall) {
    const { content, removed } = removeHookFromContent(existing);
    if (removed) {
      fs.outputFileSync(profilePath, content, 'utf8');
      return {
        shell,
        profilePath,
        snippet,
        action: 'uninstalled'
      };
    }
    return {
      shell,
      profilePath,
      snippet,
      action: 'not-installed'
    };
  }

  const { content, status } = applyHookToContent(existing, snippet);
  if (status !== 'already-installed' || !fs.pathExistsSync(profilePath)) {
    fs.outputFileSync(profilePath, content, 'utf8');
  }

  return {
    shell,
    profilePath,
    snippet,
    action: status
  };
}
