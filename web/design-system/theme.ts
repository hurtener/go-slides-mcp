/**
 * Deckard design system — theme controller.
 *
 * Implements the white-label precedence chain on the App root element:
 *   1. built-in preset      — `data-app-theme` attribute selects a token block
 *   2. client brand tokens  — inline `--app-*` overrides from the startup brand JSON
 *   3. host context vars     — applied last by the surface (Claude theme), wins
 *
 * No surface hardcodes a color; everything resolves to a `--app-*` token, so
 * any of these layers re-skins the whole UI.
 */

export type AppThemeId =
  | 'deckard-white'
  | 'deckard-dark'
  | 'midnight'
  | 'slate'
  | 'editorial-sepia';

export type ThemeMode = 'light' | 'dark';

export interface ThemeDef {
  id: AppThemeId;
  label: string;
  mode: ThemeMode;
}

/** The built-in theme presets shown in the selector, in display order. */
export const THEMES: ThemeDef[] = [
  { id: 'deckard-white', label: 'Deckard White', mode: 'light' },
  { id: 'deckard-dark', label: 'Deckard Dark', mode: 'dark' },
  { id: 'midnight', label: 'Midnight', mode: 'dark' },
  { id: 'slate', label: 'Slate', mode: 'light' },
  { id: 'editorial-sepia', label: 'Editorial Sepia', mode: 'light' },
];

export const DEFAULT_THEME: AppThemeId = 'deckard-white';

/**
 * Brand configuration delivered by the server at startup (white-label). Any
 * field is optional; an empty config leaves the built-in Deckard White intact.
 */
export interface BrandConfig {
  /** Product/brand title shown in surface chrome. */
  title?: string;
  /** The theme selected by default for this deployment. */
  defaultTheme?: AppThemeId;
  /** Per-token `--app-*` overrides (without the leading `--`), e.g. {"app-accent":"#7c3aed"}. */
  tokens?: Record<string, string>;
  /** When false, hide the theme selector (a locked single-brand deployment). */
  allowThemeSwitch?: boolean;
}

export function themeById(id: string | undefined): AppThemeId {
  return THEMES.some((t) => t.id === id) ? (id as AppThemeId) : DEFAULT_THEME;
}

/** Selects a built-in preset by setting the root's `data-app-theme`. */
export function applyTheme(root: HTMLElement, id: AppThemeId): void {
  root.setAttribute('data-app-theme', id);
}

/**
 * Applies client brand token overrides as inline custom properties. Keys may
 * be given with or without the leading `--`. Called once at startup, before
 * host vars; the inline style sits above the preset block in the cascade.
 */
export function applyBrandTokens(root: HTMLElement, tokens: Record<string, string> | undefined): void {
  if (!tokens) return;
  for (const [rawKey, value] of Object.entries(tokens)) {
    const key = rawKey.startsWith('--') ? rawKey : `--${rawKey}`;
    if (key.startsWith('--app-')) root.style.setProperty(key, value);
  }
}

/**
 * Applies the host's style variables last (highest precedence). The bridge
 * hands us a flat map; we forward only `--app-*` keys so a host theme can
 * override Deckard tokens directly, and honor a light/dark hint.
 */
export function applyHostVariables(root: HTMLElement, vars: Record<string, string> | undefined): void {
  if (!vars) return;
  for (const [key, value] of Object.entries(vars)) {
    if (key.startsWith('--app-')) root.style.setProperty(key, value);
  }
}

/** Reads a light/dark hint from host variables (color-scheme or a dy theme attr). */
export function hostModeHint(vars: Record<string, string> | undefined): ThemeMode | undefined {
  if (!vars) return undefined;
  const cs = vars['color-scheme'] ?? vars['--dy-color-scheme'];
  if (cs === 'dark' || cs === 'light') return cs;
  return undefined;
}
