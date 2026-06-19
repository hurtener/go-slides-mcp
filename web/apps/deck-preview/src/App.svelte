<!--
  deck-preview — Deckard's default inline, glanceable surface (Editorial hero +
  strip). Shows a deck as a large featured slide over a thumbnail filmstrip,
  with download / overview / edit handoffs that call the SAME agent tools.
  Never auto-opens fullscreen. Themed through the white-label --app-* chain.
-->
<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { createBridge } from 'dockyard-bridge';
  import { PageState } from 'dockyard-ui';
  import type { PageStateValue } from 'dockyard-ui';

  import SlideThumb, { type SlidePreview } from './SlideThumb.svelte';
  import ThemeSelector from '../../../design-system/ThemeSelector.svelte';
  import {
    applyTheme,
    applyBrandTokens,
    applyHostVariables,
    themeById,
    type AppThemeId,
  } from '../../../design-system/theme';
  import '../../../design-system/base.css';
  import '../../../design-system/tokens.css';

  interface Brand {
    title?: string;
    defaultTheme?: string;
    tokens?: Record<string, string>;
    allowThemeSwitch?: boolean;
  }
  interface DeckSummary { deckId: string; slug?: string; title: string; slideCount: number; soulId?: string }
  interface Payload {
    state: 'ready' | 'empty' | 'error' | 'permission' | 'loading';
    message?: string;
    brand: Brand;
    deck: DeckSummary;
    slides?: SlidePreview[];
    resourceUri?: string;
  }

  let rootEl = $state<HTMLDivElement | undefined>(undefined);
  let pageState: PageStateValue = $state('loading');
  let message = $state('Loading deck…');
  let payload = $state<Payload | null>(null);
  let selected = $state(0);
  let theme: AppThemeId = $state('deckard-white');
  let userPicked = $state(false);
  let hostVars: Record<string, string> | undefined;
  let toast = $state('');

  const bridge = createBridge({ displayModes: ['inline'] });

  const slides = $derived(payload?.slides ?? []);
  const featured = $derived(slides[selected] ?? slides[0]);
  const brandTitle = $derived(payload?.brand?.title || 'Deckard Slides');
  const allowSwitch = $derived(payload?.brand?.allowThemeSwitch !== false);

  function applyChain() {
    if (!rootEl) return;
    applyTheme(rootEl, theme); // 1. preset
    applyBrandTokens(rootEl, payload?.brand?.tokens); // 2. brand (startup)
    applyHostVariables(rootEl, hostVars); // 3. host (runtime, wins)
  }

  const offResult = bridge.onToolResult<Payload>((r) => {
    if (!r.structuredContent) {
      pageState = 'error';
      message = 'The tool returned no preview payload.';
      return;
    }
    payload = r.structuredContent;
    pageState = payload.state === 'permission' ? 'error' : (payload.state as PageStateValue);
    message = payload.message ?? '';
    if (!userPicked) theme = themeById(payload.brand?.defaultTheme);
    selected = 0;
    applyChain();
  });

  const offHost = bridge.onHostContextChanged((p) => {
    if (p.styles?.variables) {
      hostVars = p.styles.variables as Record<string, string>;
      applyChain();
    }
  });

  onMount(() => {
    applyChain();
    bridge.connect().catch((err: unknown) => {
      pageState = 'error';
      message = `Bridge handshake failed: ${(err as Error)?.message ?? err}`;
    });
  });
  onDestroy(() => { offResult(); offHost(); bridge.close(); });

  function pickTheme(id: AppThemeId) { userPicked = true; theme = id; applyChain(); }
  function flash(msg: string) { toast = msg; setTimeout(() => (toast = ''), 2600); }

  // Inline action icons (stroke = currentColor so they adopt the theme).
  const ICON_DOWNLOAD =
    '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3v12"/><path d="m7 10 5 5 5-5"/><path d="M5 21h14"/></svg>';
  const ICON_GRID =
    '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><rect x="3" y="3" width="7" height="7" rx="1.5"/><rect x="14" y="3" width="7" height="7" rx="1.5"/><rect x="3" y="14" width="7" height="7" rx="1.5"/><rect x="14" y="14" width="7" height="7" rx="1.5"/></svg>';
  const ICON_EDIT =
    '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20h9"/><path d="M16.5 3.5a2.12 2.12 0 0 1 3 3L7 19l-4 1 1-4Z"/></svg>';

  async function download() {
    if (!payload) return;
    try {
      await bridge.callTool('export_deck', { deckId: payload.deck.deckId });
      flash('Export ready — check the deck:// resource.');
    } catch (e) { flash(`Export failed: ${(e as Error)?.message ?? e}`); }
  }
  async function openOverview() {
    if (!payload) return;
    try { await bridge.callTool('get_deck_overview', { deckId: payload.deck.deckId }); }
    catch (e) { flash(`Couldn't open overview: ${(e as Error)?.message ?? e}`); }
  }
  async function editSelected() {
    if (!payload || !featured) return;
    try { await bridge.callTool('open_slide_editor', { deckId: payload.deck.deckId, slideId: featured.id }); }
    catch (e) { flash(`Couldn't open editor: ${(e as Error)?.message ?? e}`); }
  }
</script>

<div bind:this={rootEl} class="dy-root preview" data-app-theme={theme}>
  <PageState
    state={pageState}
    loadingMessage="Loading deck…"
    emptyTitle="No slides yet"
    emptyDescription={message || 'Add a slide to see it previewed here.'}
    errorTitle="Couldn't load the preview"
    errorDescription={message}
    onRetry={() => { pageState = 'loading'; bridge.connect().catch(() => {}); }}
  >
    {#if payload}
      <header class="head">
        <div class="titles">
          <h1>{payload.deck.title || 'Untitled deck'}</h1>
          <span class="meta">{payload.deck.slideCount} slide{payload.deck.slideCount === 1 ? '' : 's'}</span>
        </div>
        <div class="tools">
          <button type="button" class="icon primary" data-tip="Download .pptx" aria-label="Download .pptx" onclick={download}>{@html ICON_DOWNLOAD}</button>
          <button type="button" class="icon" data-tip="Open overview" aria-label="Open overview" onclick={openOverview}>{@html ICON_GRID}</button>
          <button type="button" class="icon" data-tip="Edit this slide" aria-label="Edit this slide" onclick={editSelected}>{@html ICON_EDIT}</button>
          {#if allowSwitch}
            <ThemeSelector current={theme} onchange={pickTheme} />
          {/if}
        </div>
      </header>

      {#if featured}
        <div class="featured"><SlideThumb slide={featured} size="featured" /></div>
      {/if}

      {#if slides.length > 1}
        <div class="rail" role="listbox" aria-label="Slides">
          {#each slides as s (s.id)}
            <SlideThumb slide={s} size="strip" selected={s.index === selected} onselect={(i) => (selected = i)} />
          {/each}
        </div>
      {/if}

      {#if toast}<p class="toast">{toast}</p>{/if}
    {/if}
  </PageState>
</div>

<style>
  .preview { padding: var(--app-space-3) var(--app-space-4); display: flex; flex-direction: column; gap: var(--app-space-3); }

  /* single-line header: title + inline count on the left, actions on the right */
  .head { display: flex; align-items: center; justify-content: space-between; gap: var(--app-space-3); }
  .titles { min-width: 0; display: flex; align-items: baseline; gap: var(--app-space-2); }
  .titles h1 {
    margin: 0;
    font-family: var(--app-font-serif);
    font-weight: var(--app-weight-medium);
    font-size: var(--app-text-lg);
    line-height: 1.2;
    color: var(--app-text);
    white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
  }
  .meta { font-size: var(--app-text-xs); color: var(--app-text-muted); white-space: nowrap; flex: 0 0 auto; }

  .tools { display: flex; align-items: center; gap: var(--app-space-2); flex: 0 0 auto; }

  /* icon action buttons with a CSS tooltip */
  .icon {
    position: relative;
    width: 30px; height: 30px;
    display: grid; place-items: center;
    border: 1px solid var(--app-border);
    border-radius: var(--app-radius-md);
    background: var(--app-surface);
    color: var(--app-text);
    cursor: pointer;
    transition: border-color var(--app-dur) var(--app-ease), background var(--app-dur) var(--app-ease), color var(--app-dur) var(--app-ease);
  }
  .icon:hover { border-color: var(--app-accent); color: var(--app-accent-text); }
  .icon:focus-visible { outline: 2px solid var(--app-accent); outline-offset: 2px; }
  .icon.primary { background: var(--app-accent); color: #fff; border-color: transparent; }
  .icon.primary:hover { background: var(--app-accent-hover); color: #fff; }
  .icon[data-tip]::after {
    content: attr(data-tip);
    position: absolute; bottom: calc(100% + 7px); left: 50%; transform: translateX(-50%);
    background: var(--app-text); color: var(--app-bg);
    font-size: var(--app-text-xs); white-space: nowrap;
    padding: 4px 8px; border-radius: var(--app-radius-sm);
    opacity: 0; pointer-events: none; transition: opacity var(--app-dur-fast) var(--app-ease);
    z-index: 30;
  }
  .icon[data-tip]:hover::after, .icon[data-tip]:focus-visible::after { opacity: 1; }

  .featured { width: 100%; }
  .featured :global(.frame) { box-shadow: var(--app-shadow-lg); }

  .rail {
    display: flex; gap: var(--app-space-2);
    justify-content: safe center;
    overflow-x: auto;
    padding: 0 2px 2px;
    scrollbar-width: thin;
  }

  .toast { margin: 0; font-size: var(--app-text-xs); color: var(--app-text-muted); text-align: center; }
</style>
