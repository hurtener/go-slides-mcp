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
  import ThemeSelector from './ThemeSelector.svelte';
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
          <p class="meta">{payload.deck.slideCount} slide{payload.deck.slideCount === 1 ? '' : 's'} · {brandTitle}</p>
        </div>
        {#if allowSwitch}
          <ThemeSelector current={theme} onchange={pickTheme} />
        {/if}
      </header>

      <div class="stage">
        {#if featured}
          <div class="featured-wrap">
            <SlideThumb slide={featured} size="featured" />
          </div>
        {/if}
        <div class="actions">
          <button type="button" class="btn primary" onclick={download}>↓ Download .pptx</button>
          <button type="button" class="btn" onclick={openOverview}>▦ Open overview</button>
          <button type="button" class="btn" onclick={editSelected}>✎ Edit this slide</button>
          {#if toast}<p class="toast">{toast}</p>{/if}
        </div>
      </div>

      {#if slides.length > 1}
        <div class="strip" role="listbox" aria-label="Slides">
          {#each slides as s (s.id)}
            <SlideThumb slide={s} size="strip" selected={s.index === selected} onselect={(i) => (selected = i)} />
          {/each}
        </div>
      {/if}
    {/if}
  </PageState>
</div>

<style>
  .preview { padding: var(--app-space-5); display: flex; flex-direction: column; gap: var(--app-space-4); }
  .head { display: flex; align-items: flex-start; justify-content: space-between; gap: var(--app-space-4); }
  .titles h1 {
    margin: 0;
    font-family: var(--app-font-serif);
    font-weight: var(--app-weight-medium);
    font-size: var(--app-text-2xl);
    line-height: 1.1;
    color: var(--app-text);
  }
  .meta { margin: 4px 0 0; font-size: var(--app-text-sm); color: var(--app-text-muted); }

  .stage { display: grid; grid-template-columns: 1fr 200px; gap: var(--app-space-5); align-items: start; }
  .featured-wrap { box-shadow: var(--app-shadow-lg); border-radius: var(--app-radius-lg); overflow: hidden; }
  .actions { display: flex; flex-direction: column; gap: var(--app-space-2); }

  .btn {
    display: inline-flex; align-items: center; gap: var(--app-space-2);
    padding: 9px 14px; border-radius: var(--app-radius-md);
    border: 1px solid var(--app-border);
    background: var(--app-surface); color: var(--app-text);
    font-size: var(--app-text-sm); cursor: pointer; text-align: left;
    transition: border-color var(--app-dur) var(--app-ease), background var(--app-dur) var(--app-ease);
  }
  .btn:hover { border-color: var(--app-accent); background: var(--app-surface-raised); }
  .btn:focus-visible { outline: 2px solid var(--app-accent); outline-offset: 2px; }
  .btn.primary { background: var(--app-accent); color: #fff; border-color: transparent; }
  .btn.primary:hover { background: var(--app-accent-hover); }
  .toast { margin: var(--app-space-2) 0 0; font-size: var(--app-text-xs); color: var(--app-text-muted); }

  .strip {
    display: flex; gap: var(--app-space-3);
    overflow-x: auto; padding: var(--app-space-2) 2px var(--app-space-3);
    scrollbar-width: thin;
  }

  @media (max-width: 560px) {
    .stage { grid-template-columns: 1fr; }
  }
</style>
