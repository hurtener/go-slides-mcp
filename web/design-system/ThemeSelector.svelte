<!--
  ThemeSelector — the multi-theme picker. Lists the built-in presets; selecting
  one re-skins the surface live (preset layer of the white-label chain). Hidden
  when the brand locks the theme (allowThemeSwitch=false).
-->
<script lang="ts">
  import { THEMES, type AppThemeId } from './theme';

  let {
    current,
    onchange,
  }: { current: AppThemeId; onchange: (id: AppThemeId) => void } = $props();

  let open = $state(false);
  const label = $derived(THEMES.find((t) => t.id === current)?.label ?? 'Theme');

  function pick(id: AppThemeId) {
    open = false;
    onchange(id);
  }
</script>

<div class="ts">
  <button type="button" class="trigger" aria-haspopup="listbox" aria-expanded={open} onclick={() => (open = !open)}>
    <span class="swatch" data-app-theme={current}></span>
    <span class="label">{label}</span>
    <span class="caret" class:open>▾</span>
  </button>
  {#if open}
    <ul class="menu" role="listbox" aria-label="Theme">
      {#each THEMES as t (t.id)}
        <li>
          <button type="button" role="option" aria-selected={t.id === current} class:active={t.id === current} onclick={() => pick(t.id)}>
            <span class="swatch" data-app-theme={t.id}></span>
            <span>{t.label}</span>
            {#if t.id === current}<span class="check">✓</span>{/if}
          </button>
        </li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  .ts { position: relative; }
  .trigger {
    display: inline-flex; align-items: center; gap: var(--app-space-2);
    padding: 6px 10px;
    border: 1px solid var(--app-border);
    border-radius: var(--app-radius-pill);
    background: var(--app-surface);
    color: var(--app-text);
    font-size: var(--app-text-sm);
    cursor: pointer;
    transition: border-color var(--app-dur) var(--app-ease);
  }
  .trigger:hover { border-color: var(--app-accent); }
  .trigger:focus-visible { outline: 2px solid var(--app-accent); outline-offset: 2px; }
  .caret { color: var(--app-text-muted); transition: transform var(--app-dur) var(--app-ease); }
  .caret.open { transform: rotate(180deg); }
  /* a swatch previews a theme by adopting its tokens, then showing paper+accent */
  .swatch {
    width: 14px; height: 14px; border-radius: 4px; flex: 0 0 auto;
    background: var(--app-paper);
    border: 1px solid var(--app-border-strong);
    box-shadow: inset -5px 0 0 var(--app-accent);
  }
  .menu {
    position: absolute; right: 0; top: calc(100% + 6px);
    z-index: 20; margin: 0; padding: 4px; list-style: none;
    min-width: 190px;
    background: var(--app-surface);
    border: 1px solid var(--app-border);
    border-radius: var(--app-radius-md);
    box-shadow: var(--app-shadow-md);
  }
  .menu li { margin: 0; }
  .menu button {
    width: 100%; display: flex; align-items: center; gap: var(--app-space-2);
    padding: 7px 8px; border: 0; border-radius: var(--app-radius-sm);
    background: transparent; color: var(--app-text);
    font-size: var(--app-text-sm); text-align: left; cursor: pointer;
  }
  .menu button:hover { background: var(--app-surface-raised); }
  .menu button.active { color: var(--app-accent-text); }
  .check { margin-left: auto; color: var(--app-accent-text); }
</style>
