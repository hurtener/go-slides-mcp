<!--
  slide-editor — opt-in single-slide editor (ui://app, tight PiP/phone widths).
  Lists the slide's nodes; text-bearing nodes get inline fields wired to the
  IR-path edit tools (patch_slide_text for RichText, edit_slide_field for string
  fields). Nodes reorder (move_slide_node), duplicate (duplicate_slide_node),
  and remove behind an ImpactModal (remove_slide_node). "← Back to deck" returns
  to the overview. Every edit calls the SAME agent tool. White-label themed.
-->
<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { createBridge } from 'dockyard-bridge';
  import { PageState } from 'dockyard-ui';
  import type { PageStateValue } from 'dockyard-ui';

  import ImpactModal from '../../../design-system/ImpactModal.svelte';
  import ThemeSelector from '../../../design-system/ThemeSelector.svelte';
  import {
    applyTheme, applyBrandTokens, applyHostVariables, themeById, type AppThemeId,
  } from '../../../design-system/theme';
  import '../../../design-system/base.css';
  import '../../../design-system/tokens.css';

  interface Brand { title?: string; defaultTheme?: string; tokens?: Record<string, string>; allowThemeSwitch?: boolean }
  type Node = Record<string, unknown> & { kind: string };
  interface Slide { id?: string; layout?: string; nodes?: Node[] }
  interface Validation { ok: boolean; issues?: string[] }
  interface Payload {
    state?: string; message?: string; brand: Brand;
    deckId: string; slideId: string; ir: Slide; soulId?: string; validation?: Validation;
  }
  interface Field { label: string; field: string; path: unknown[]; rich: boolean; value: string }

  let rootEl = $state<HTMLDivElement | undefined>(undefined);
  let pageState: PageStateValue = $state('loading');
  let message = $state('Loading slide…');
  let payload = $state<Payload | null>(null);
  let nodes = $state<Node[]>([]);
  let validation = $state<Validation | null>(null);
  let theme: AppThemeId = $state('deckard-white');
  let userPicked = $state(false);
  let hostVars: Record<string, string> | undefined;
  let toast = $state('');
  let pendingRemove = $state<number | null>(null);

  const bridge = createBridge({ displayModes: ['inline'] });
  const allowSwitch = $derived(payload?.brand?.allowThemeSwitch !== false);
  const deckId = $derived(payload?.deckId ?? '');
  const slideId = $derived(payload?.slideId ?? '');

  function applyChain() {
    if (!rootEl) return;
    applyTheme(rootEl, theme);
    applyBrandTokens(rootEl, payload?.brand?.tokens);
    applyHostVariables(rootEl, hostVars);
  }

  const offResult = bridge.onToolResult<Payload>((r) => {
    if (!r.structuredContent) { pageState = 'error'; message = 'The tool returned no payload.'; return; }
    payload = r.structuredContent;
    nodes = [...(payload.ir?.nodes ?? [])];
    validation = payload.validation ?? null;
    const st = payload.state ?? 'ready';
    pageState = st === 'permission' ? 'error' : (st as PageStateValue);
    message = payload.message ?? '';
    if (!userPicked) theme = themeById(payload.brand?.defaultTheme);
    applyChain();
  });
  const offHost = bridge.onHostContextChanged((p) => {
    if (p.styles?.variables) { hostVars = p.styles.variables as Record<string, string>; applyChain(); }
  });

  onMount(() => {
    applyChain();
    bridge.connect().catch((e: unknown) => { pageState = 'error'; message = `Bridge handshake failed: ${(e as Error)?.message ?? e}`; });
  });
  onDestroy(() => { offResult(); offHost(); bridge.close(); });

  function pickTheme(id: AppThemeId) { userPicked = true; theme = id; applyChain(); }
  function flash(m: string) { toast = m; setTimeout(() => (toast = ''), 2800); }

  function rtText(v: unknown): string {
    return Array.isArray(v) ? v.map((r) => (r && typeof r === 'object' ? String((r as { text?: string }).text ?? '') : '')).join('') : '';
  }
  function kindLabel(k: string): string { return k.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase()); }

  // editable fields for a node at top-level index i (supported kinds only).
  function fieldsFor(n: Node, i: number): Field[] {
    const f: Field[] = [];
    // IR paths are "nodes"-prefixed (what the edit tools resolve): ["nodes", i, ...].
    const str = (label: string, field: string) => f.push({ label, field, path: ['nodes', i], rich: false, value: String((n[field] as string) ?? '') });
    const rich = (label: string, field: string) => f.push({ label, field, path: ['nodes', i], rich: true, value: rtText(n[field]) });
    switch (n.kind) {
      case 'hero': str('Eyebrow', 'eyebrow'); str('Title', 'title'); str('Subtitle', 'subtitle'); break;
      case 'heading': rich('Heading', 'text'); break;
      case 'callout': str('Title', 'title'); rich('Body', 'body'); break;
      case 'quote': rich('Quote', 'text'); str('Attribution', 'attribution'); break;
      case 'chip': str('Label', 'label'); break;
      case 'section_divider': str('Eyebrow', 'eyebrow'); str('Label', 'label'); break;
      case 'arrow': str('Label', 'label'); break;
      case 'list': {
        const items = (n.items as Array<{ text?: unknown }>) ?? [];
        items.forEach((it, j) => f.push({ label: `Item ${j + 1}`, field: 'text', path: ['nodes', i, 'items', j], rich: true, value: rtText(it.text) }));
        break;
      }
    }
    return f;
  }
  function previewFor(n: Node): string {
    switch (n.kind) {
      case 'prose': { const p = (n.paragraphs as unknown[]) ?? []; return p.length ? rtText(p[0]) : 'Paragraph text'; }
      case 'chart': return 'Chart image (edit via the agent)';
      case 'code_block': return `Code${n.language ? ` · ${n.language}` : ''}`;
      case 'image': return 'Image';
      case 'table': return 'Table';
      case 'divider': return 'Divider';
      case 'flow': return 'Process flow';
      case 'two_column': return 'Two-column layout';
      case 'grid': return 'Card grid';
      case 'card': case 'card_section': return String((n.header as string) ?? 'Card');
      default: return '';
    }
  }

  function save(fl: Field, value: string) {
    if (value === fl.value) return;
    fl.value = value;
    const base = { deckId, slideId, path: fl.path, field: fl.field };
    const call = fl.rich
      ? bridge.callTool('patch_slide_text', { ...base, text: value })
      : bridge.callTool('edit_slide_field', { ...base, value });
    void call
      .then((r: unknown) => { const v = (r as { structuredContent?: { validation?: Validation } })?.structuredContent?.validation; if (v) validation = v; })
      .catch((e: unknown) => flash(`Edit failed: ${(e as Error)?.message ?? e}`));
  }

  function moveNode(i: number, dir: -1 | 1) {
    const j = i + dir;
    if (j < 0 || j >= nodes.length) return;
    const next = [...nodes];
    [next[i], next[j]] = [next[j], next[i]];
    nodes = next;
    void bridge.callTool('move_slide_node', { deckId, slideId, from: ['nodes', j], to: ['nodes', i] })
      .catch((e: unknown) => flash(`Move failed: ${(e as Error)?.message ?? e}`));
  }
  function duplicateNode(i: number) {
    void bridge.callTool('duplicate_slide_node', { deckId, slideId, path: ['nodes', i] })
      .then(() => bridge.callTool('open_slide_editor', { deckId, slideId }))
      .catch((e: unknown) => flash(`Duplicate failed: ${(e as Error)?.message ?? e}`));
  }
  function confirmRemove() {
    const i = pendingRemove; pendingRemove = null;
    if (i === null) return;
    nodes = nodes.filter((_, idx) => idx !== i);
    void bridge.callTool('remove_slide_node', { deckId, slideId, path: ['nodes', i] })
      .catch((e: unknown) => flash(`Remove failed: ${(e as Error)?.message ?? e}`));
  }
  function back() {
    void bridge.callTool('get_deck_overview', { deckId }).catch(() => flash('Couldn’t return to the deck.'));
  }
</script>

<div bind:this={rootEl} class="dy-root editor" data-app-theme={theme}>
  <PageState
    state={pageState}
    loadingMessage="Loading slide…"
    emptyTitle="Empty slide"
    emptyDescription={message || 'This slide has no nodes yet.'}
    errorTitle="Couldn't open the editor"
    errorDescription={message}
    onRetry={() => { pageState = 'loading'; bridge.connect().catch(() => {}); }}
  >
    {#if payload}
      <header class="head">
        <button type="button" class="back" onclick={back} aria-label="Back to deck">‹ Deck</button>
        <span class="title">Edit slide</span>
        {#if allowSwitch}<ThemeSelector current={theme} onchange={pickTheme} />{/if}
      </header>

      {#if validation && !validation.ok}
        <div class="banner" role="status">
          <strong>{validation.issues?.length ?? 0} issue{(validation.issues?.length ?? 0) === 1 ? '' : 's'}</strong>
          {#if validation.issues?.length}<span>· {validation.issues[0]}</span>{/if}
        </div>
      {/if}

      <div class="nodes">
        {#each nodes as n, i (i)}
          {@const fields = fieldsFor(n, i)}
          <section class="node">
            <div class="nhead">
              <span class="chip">{kindLabel(n.kind)}</span>
              <span class="nacts">
                <button type="button" class="mini" data-tip="Move up" aria-label="Move up" disabled={i === 0} onclick={() => moveNode(i, -1)}>↑</button>
                <button type="button" class="mini" data-tip="Move down" aria-label="Move down" disabled={i === nodes.length - 1} onclick={() => moveNode(i, 1)}>↓</button>
                <button type="button" class="mini" data-tip="Duplicate" aria-label="Duplicate node" onclick={() => duplicateNode(i)}>⧉</button>
                <button type="button" class="mini danger" data-tip="Remove" aria-label="Remove node" onclick={() => (pendingRemove = i)}>🗑</button>
              </span>
            </div>
            {#if fields.length}
              <div class="fields">
                {#each fields as fl (fl.field + fl.path.join('.'))}
                  <label class="field">
                    <span class="flabel">{fl.label}</span>
                    {#if fl.rich}
                      <textarea rows="2" value={fl.value} onblur={(e) => save(fl, (e.currentTarget as HTMLTextAreaElement).value)}></textarea>
                    {:else}
                      <input type="text" value={fl.value} onblur={(e) => save(fl, (e.currentTarget as HTMLInputElement).value)} />
                    {/if}
                  </label>
                {/each}
              </div>
            {:else}
              <p class="preview">{previewFor(n)}</p>
            {/if}
          </section>
        {/each}
        {#if nodes.length === 0}
          <p class="preview empty">This slide has no nodes.</p>
        {/if}
      </div>

      {#if toast}<p class="toast">{toast}</p>{/if}
    {/if}
  </PageState>

  <ImpactModal
    open={pendingRemove !== null}
    title="Remove this node?"
    message="The node will be removed from the slide. This can’t be undone from here."
    confirmLabel="Remove node"
    danger
    onconfirm={confirmRemove}
    oncancel={() => (pendingRemove = null)}
  />
</div>

<style>
  .editor { padding: var(--app-space-3) var(--app-space-4); display: flex; flex-direction: column; gap: var(--app-space-3); }
  .head { display: flex; align-items: center; justify-content: space-between; gap: var(--app-space-2); }
  .back {
    border: 1px solid var(--app-border); background: var(--app-surface); color: var(--app-text);
    border-radius: var(--app-radius-md); padding: 5px 10px; font-size: var(--app-text-sm); cursor: pointer; flex: 0 0 auto;
  }
  .back:hover { border-color: var(--app-accent); color: var(--app-accent-text); }
  .back:focus-visible { outline: 2px solid var(--app-accent); outline-offset: 2px; }
  .title { flex: 1; min-width: 0; font-family: var(--app-font-serif); font-weight: var(--app-weight-medium); font-size: var(--app-text-md); color: var(--app-text); text-align: center; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

  .banner {
    font-size: var(--app-text-xs); color: var(--app-warning);
    background: color-mix(in srgb, var(--app-warning) 12%, transparent);
    border: 1px solid color-mix(in srgb, var(--app-warning) 35%, transparent);
    border-radius: var(--app-radius-sm); padding: 6px 10px;
  }
  .banner span { color: var(--app-text-muted); }

  .nodes { display: flex; flex-direction: column; gap: var(--app-space-2); }
  .node { border: 1px solid var(--app-border); border-radius: var(--app-radius-md); background: var(--app-surface); padding: var(--app-space-2) var(--app-space-3) var(--app-space-3); }
  .nhead { display: flex; align-items: center; justify-content: space-between; }
  .chip { font-size: var(--app-text-xs); color: var(--app-accent-text); background: var(--app-accent-soft); border-radius: var(--app-radius-pill); padding: 1px 8px; }
  .nacts { display: flex; gap: 1px; }
  .mini { position: relative; width: 24px; height: 24px; display: grid; place-items: center; border: 0; border-radius: var(--app-radius-sm); background: transparent; color: var(--app-text-muted); cursor: pointer; font-size: 12px; }
  .mini:hover:not(:disabled) { background: var(--app-surface-raised); color: var(--app-text); }
  .mini.danger:hover:not(:disabled) { color: var(--app-danger); }
  .mini:disabled { opacity: 0.3; cursor: default; }
  .mini:focus-visible { outline: 2px solid var(--app-accent); outline-offset: 1px; }
  .mini[data-tip]::after { content: attr(data-tip); position: absolute; bottom: calc(100% + 5px); left: 50%; transform: translateX(-50%); background: var(--app-text); color: var(--app-bg); font-size: var(--app-text-xs); white-space: nowrap; padding: 3px 6px; border-radius: var(--app-radius-sm); opacity: 0; pointer-events: none; transition: opacity var(--app-dur-fast) var(--app-ease); z-index: 30; }
  .mini[data-tip]:hover::after { opacity: 1; }

  .fields { display: flex; flex-direction: column; gap: var(--app-space-2); margin-top: var(--app-space-2); }
  .field { display: flex; flex-direction: column; gap: 3px; }
  .flabel { font-size: var(--app-text-xs); color: var(--app-text-muted); }
  input, textarea {
    font: inherit; font-size: var(--app-text-sm); color: var(--app-text);
    background: var(--app-bg); border: 1px solid var(--app-border); border-radius: var(--app-radius-sm);
    padding: 6px 8px; width: 100%; resize: vertical;
  }
  input:focus, textarea:focus { outline: none; border-color: var(--app-accent); box-shadow: 0 0 0 2px var(--app-accent-soft); }
  .preview { margin: var(--app-space-2) 0 0; font-size: var(--app-text-sm); color: var(--app-text-muted); }
  .preview.empty { text-align: center; }
  .toast { margin: 0; font-size: var(--app-text-xs); color: var(--app-text-muted); text-align: center; }
</style>
