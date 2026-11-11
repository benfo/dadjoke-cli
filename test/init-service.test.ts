import { describe, it, expect } from 'vitest';
import fs from 'fs-extra';
import path from 'node:path';
import os from 'node:os';
import {
  detectShell,
  resolveProfilePath,
  buildHookSnippet,
  configureShellHook,
  HOOK_START_MARKER,
  HOOK_END_MARKER
} from '../src/init-service.js';

describe('init-service', () => {
  it('detects shell from environment and platform', () => {
    expect(detectShell('linux', { SHELL: '/bin/zsh' })).toBe('zsh');
    expect(detectShell('linux', { SHELL: '/bin/bash' })).toBe('bash');
    expect(detectShell('win32', { SHELL: '' })).toBe('powershell');
    expect(detectShell('darwin', { SHELL: '/usr/local/bin/pwsh' })).toBe('powershell');
  });

  it('resolves profile paths for powershell, bash, and zsh', () => {
    const fakeHome = '/home/tester';
    expect(
      resolveProfilePath('zsh', { homeDir: fakeHome, platform: 'linux', env: {} })
    ).toBe(path.join(fakeHome, '.zshrc'));

    expect(
      resolveProfilePath('bash', { homeDir: fakeHome, platform: 'linux', env: {} })
    ).toBe(path.join(fakeHome, '.bashrc'));

    expect(
      resolveProfilePath('powershell', {
        homeDir: fakeHome,
        platform: 'win32',
        env: { PROFILE: 'C:\\Users\\tester\\CustomProfile.ps1' }
      })
    ).toBe('C:\\Users\\tester\\CustomProfile.ps1');
  });

  it('builds valid hook snippets for powershell and bash/zsh', () => {
    const psSnippet = buildHookSnippet('powershell', 'C:\\app\\dist\\cli.js');
    expect(psSnippet).toContain(HOOK_START_MARKER);
    expect(psSnippet).toContain('Get-Command dadjoke');
    expect(psSnippet).toContain(HOOK_END_MARKER);

    const bashSnippet = buildHookSnippet('bash', '/app/dist/cli.js');
    expect(bashSnippet).toContain('command -v dadjoke');
    expect(bashSnippet).toContain('/app/dist/cli.js');
  });

  it('installs, idempotently checks, updates, and uninstalls shell hook', () => {
    const tmpDir = path.join(os.tmpdir(), `dadjoke-init-test-${Date.now()}`);
    const profileFile = path.join(tmpDir, 'Microsoft.PowerShell_profile.ps1');

    try {
      fs.outputFileSync(profileFile, '# Existing user alias\nSet-Alias ll Get-ChildItem\n', 'utf8');

      // 1. First install
      const res1 = configureShellHook({
        shell: 'powershell',
        profilePath: profileFile,
        cliScriptPath: 'C:\\path1\\cli.js'
      });
      expect(res1.action).toBe('installed');
      const content1 = fs.readFileSync(profileFile, 'utf8');
      expect(content1).toContain('Set-Alias ll Get-ChildItem');
      expect(content1).toContain('C:\\path1\\cli.js');

      // 2. Second install with identical config is a no-op
      const res2 = configureShellHook({
        shell: 'powershell',
        profilePath: profileFile,
        cliScriptPath: 'C:\\path1\\cli.js'
      });
      expect(res2.action).toBe('already-installed');

      // 3. Install with updated path replaces block in-place without duplicating
      const res3 = configureShellHook({
        shell: 'powershell',
        profilePath: profileFile,
        cliScriptPath: 'C:\\path2\\cli.js'
      });
      expect(res3.action).toBe('updated');
      const content3 = fs.readFileSync(profileFile, 'utf8');
      expect(content3).not.toContain('C:\\path1\\cli.js');
      expect(content3).toContain('C:\\path2\\cli.js');
      expect(content3.match(new RegExp(HOOK_START_MARKER, 'g'))?.length).toBe(1);

      // 4. Uninstall removes block and preserves user lines
      const res4 = configureShellHook({
        shell: 'powershell',
        profilePath: profileFile,
        uninstall: true
      });
      expect(res4.action).toBe('uninstalled');
      const content4 = fs.readFileSync(profileFile, 'utf8');
      expect(content4).toContain('Set-Alias ll Get-ChildItem');
      expect(content4).not.toContain(HOOK_START_MARKER);
    } finally {
      fs.removeSync(tmpDir);
    }
  });

  it('supports dry-run without creating or modifying files', () => {
    const tmpDir = path.join(os.tmpdir(), `dadjoke-dryrun-${Date.now()}`);
    const profileFile = path.join(tmpDir, '.zshrc');

    const res = configureShellHook({
      shell: 'zsh',
      profilePath: profileFile,
      dryRun: true
    });

    expect(res.action).toBe('dry-run');
    expect(fs.pathExistsSync(profileFile)).toBe(false);
  });
});
