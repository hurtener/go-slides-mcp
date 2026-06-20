<!--
  SlideThumb — renders one SlidePreview as a miniature slide, natively from the
  IR node descriptors (no server-side raster). The thumbnail sits on the deck
  "paper" (--app-paper). Two sizes: 'featured' (the large full-width hero) and
  'strip' (the small filmstrip items, with the slide number beneath).
-->
<script lang="ts">
  export interface ThumbNode {
    kind: string;
    text?: string;
    detail?: string;
    count?: number;
    accent?: boolean;
    children?: ThumbNode[];
    items?: string[];
  }
  export interface SlidePreview {
    id: string;
    index: number;
    layout: string;
    title?: string;
    nodes?: ThumbNode[];
  }

  let {
    slide,
    size = 'strip',
    selected = false,
    onselect,
  }: {
    slide: SlidePreview;
    size?: 'featured' | 'strip';
    selected?: boolean;
    onselect?: (index: number) => void;
  } = $props();

  const isCover = $derived(slide.layout === 'cover');
  const nodes = $derived(slide.nodes ?? []);

  const cap = (n: number, max: number) => Math.min(n || 0, max);
</script>

{#snippet nodeBlock(node: ThumbNode)}
  {#if node.kind === 'hero'}
    <div class="n-hero">
      {#if node.detail}<span class="eyebrow">{node.detail}</span>{/if}
      <span class="hero-title">{node.text}</span>
    </div>
  {:else if node.kind === 'heading' || node.kind === 'section_divider'}
    <span class="n-heading">{node.text}</span>
  {:else if node.kind === 'prose'}
    {#if node.text}
      <span class="n-text">{node.text}</span>
    {:else}
      <span class="n-line"></span><span class="n-line short"></span>
    {/if}
  {:else if node.kind === 'list'}
    {#if node.items && node.items.length}
      {#each node.items.slice(0, 4) as item, i (i)}
        <span class="n-bullet"><i></i>{#if item}<span class="n-bullet-text">{item}</span>{:else}<b></b>{/if}</span>
      {/each}
    {:else}
      {#each Array(cap(node.count ?? 3, 4) || 3) as _, i (i)}
        <span class="n-bullet"><i></i><b></b></span>
      {/each}
    {/if}
  {:else if node.kind === 'callout'}
    <div class="n-callout" class:accent={node.accent}>
      <span class="n-callout-title">{node.text || 'Callout'}</span>
      {#if node.detail}<span class="n-callout-body">{node.detail}</span>{/if}
    </div>
  {:else if node.kind === 'quote'}
    <div class="n-quote">{node.text}</div>
  {:else if node.kind === 'chart'}
    <div class="n-chart"><i style="height:40%"></i><i style="height:75%"></i><i style="height:55%"></i><i style="height:90%"></i></div>
  {:else if node.kind === 'code_block'}
    <div class="n-code"><i></i><i class="ind"></i><i class="ind"></i><i></i></div>
  {:else if node.kind === 'table'}
    <div class="n-table"><span></span><span></span><span></span><span></span><span></span><span></span></div>
  {:else if node.kind === 'image'}
    <div class="n-image"></div>
  {:else if node.kind === 'divider'}
    <span class="n-divider"></span>
  {:else if node.kind === 'chip'}
    <span class="n-chip" class:accent={node.accent}>{node.text || ''}</span>
  {:else if node.kind === 'flow'}
    <div class="n-flow">
      {#if node.items && node.items.length}
        {#each node.items.slice(0, 4) as label, i (i)}
          <span class="n-flow-step">{label}</span>{#if i < cap(node.items.length, 4) - 1}<em>›</em>{/if}
        {/each}
      {:else}
        {#each Array(cap(node.count ?? 3, 4) || 3) as _, i (i)}
          <span class="n-flow-step"></span>{#if i < (cap(node.count ?? 3, 4) || 3) - 1}<em>›</em>{/if}
        {/each}
      {/if}
    </div>
  {:else if node.kind === 'two_column'}
    {#if node.children && node.children.length}
      <div class="n-cols">
        {#each node.children as child, i (i)}
          <div class="n-col">{@render nodeBlock(child)}</div>
        {/each}
      </div>
    {:else}
      <div class="n-cols"><div></div><div></div></div>
    {/if}
  {:else if node.kind === 'grid'}
    {#if node.children && node.children.length}
      <div class="n-grid">
        {#each node.children as child, i (i)}
          <div class="n-cell">{@render nodeBlock(child)}</div>
        {/each}
      </div>
    {:else}
      <div class="n-grid">
        {#each Array(cap(node.count ?? 4, 6) || 4) as _, i (i)}<div></div>{/each}
      </div>
    {/if}
  {:else if node.kind === 'card' || node.kind === 'card_section'}
    <div class="n-card">
      {#if node.detail}<span class="n-card-eyebrow">{node.detail}</span>{/if}
      {#if node.text}<span class="n-card-header">{node.text}</span>{/if}
      {#if node.children && node.children.length}
        {#each node.children as child, i (i)}{@render nodeBlock(child)}{/each}
      {:else if !node.text}
        <span class="n-line"></span>
      {/if}
    </div>
  {:else}
    <span class="n-line"></span>
  {/if}
{/snippet}

<button
  type="button"
  class="thumb {size}"
  class:selected
  aria-label={`Slide ${slide.index + 1}${slide.title ? `: ${slide.title}` : ''}`}
  onclick={() => onselect?.(slide.index)}
>
  <div class="frame" class:selected>
    <div class="paper" class:cover={isCover}>
      {#each nodes as node, i (node.kind + (node.text ?? '') + i)}
        {@render nodeBlock(node)}
      {/each}
      {#if nodes.length === 0}
        <span class="n-heading muted">{slide.title || 'Empty slide'}</span>
      {/if}
    </div>
  </div>
  {#if size === 'strip'}
    <span class="num">{slide.index + 1}</span>
  {/if}
</button>

<style>
  .thumb {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 3px;
    padding: 0;
    border: 0;
    background: transparent;
    cursor: pointer;
  }
  .thumb.featured { width: 100%; }
  .thumb.strip { width: 84px; flex: 0 0 auto; }
  .thumb:focus-visible { outline: none; }
  .thumb:focus-visible .frame { outline: 2px solid var(--app-accent); outline-offset: 2px; }

  .frame {
    width: 100%;
    border: 1px solid var(--app-border);
    border-radius: var(--app-radius-md);
    background: var(--app-surface);
    overflow: hidden;
    transition: border-color var(--app-dur) var(--app-ease),
      box-shadow var(--app-dur) var(--app-ease);
  }
  .thumb:hover .frame { border-color: var(--app-accent); box-shadow: var(--app-shadow-sm); }
  .frame.selected { border-color: var(--app-accent); box-shadow: var(--app-shadow-md); }
  .thumb.featured .frame { border-radius: var(--app-radius-lg); }

  /* 16:9 paper */
  .paper {
    aspect-ratio: 16 / 9;
    background: var(--app-paper);
    color: var(--app-paper-ink);
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: 3%;
    padding: 7% 8%;
    overflow: hidden;
  }
  .paper.cover { justify-content: center; align-items: flex-start; }
  .thumb.featured .paper { gap: 2.4%; padding: 5.5% 6.5%; }

  .n-hero { display: flex; flex-direction: column; gap: 4px; }
  .eyebrow {
    font-size: var(--app-text-xs);
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--app-accent-text);
    opacity: 0.85;
  }
  .hero-title {
    font-family: var(--app-font-serif);
    font-weight: var(--app-weight-medium);
    line-height: 1.1;
  }
  .n-heading {
    font-family: var(--app-font-serif);
    font-weight: var(--app-weight-medium);
    line-height: 1.15;
    color: var(--app-paper-ink);
  }
  .n-heading.muted { color: var(--app-text-subtle); }
  .n-line { height: 0.5em; border-radius: 2px; background: color-mix(in srgb, var(--app-paper-ink) 22%, transparent); }
  .n-line.short { width: 62%; }
  .n-text {
    line-height: 1.3;
    color: color-mix(in srgb, var(--app-paper-ink) 78%, transparent);
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    line-clamp: 3;
    -webkit-box-orient: vertical;
  }
  .n-bullet { display: flex; align-items: center; gap: 5px; }
  .n-bullet i { width: 5px; height: 5px; border-radius: 50%; background: var(--app-accent); flex: 0 0 auto; }
  .n-bullet b { height: 0.42em; flex: 1; border-radius: 2px; background: color-mix(in srgb, var(--app-paper-ink) 20%, transparent); }
  .n-bullet-text {
    flex: 1; min-width: 0;
    color: color-mix(in srgb, var(--app-paper-ink) 78%, transparent);
    line-height: 1.2;
    overflow: hidden; white-space: nowrap; text-overflow: ellipsis;
  }
  .n-callout {
    display: flex; flex-direction: column; gap: 2px;
    border-left: 3px solid var(--app-accent);
    background: var(--app-accent-soft);
    border-radius: var(--app-radius-sm);
    padding: 4% 5%;
    font-size: var(--app-text-xs);
    color: var(--app-paper-ink);
    overflow: hidden;
  }
  .n-callout-title {
    font-weight: var(--app-weight-medium);
    overflow: hidden; white-space: nowrap; text-overflow: ellipsis;
  }
  .n-callout-body {
    color: color-mix(in srgb, var(--app-paper-ink) 72%, transparent);
    line-height: 1.25;
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
  }
  .n-quote {
    font-family: var(--app-font-serif);
    font-style: italic;
    border-left: 3px solid var(--app-border-strong);
    padding-left: 6%;
    color: var(--app-text-muted);
    overflow: hidden;
  }
  .n-chart { display: flex; align-items: flex-end; gap: 8%; height: 50%; }
  .n-chart i { flex: 1; background: var(--app-accent); border-radius: 2px 2px 0 0; opacity: 0.9; }
  .n-code {
    background: #2b2723; border-radius: var(--app-radius-sm); padding: 5%;
    display: flex; flex-direction: column; gap: 5px; flex: 1; justify-content: center;
  }
  .n-code i { height: 0.4em; border-radius: 2px; background: rgba(250, 247, 242, 0.55); width: 80%; }
  .n-code i.ind { width: 60%; margin-left: 12%; }
  .n-table { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 3px; }
  .n-table span { height: 0.7em; background: color-mix(in srgb, var(--app-paper-ink) 12%, transparent); border-radius: 2px; }
  .n-table span:nth-child(-n+3) { background: var(--app-accent-soft); }
  .n-image { flex: 1; min-height: 30%; background: color-mix(in srgb, var(--app-paper-ink) 10%, transparent); border-radius: var(--app-radius-sm); }
  .n-divider { height: 1px; background: var(--app-border-strong); }
  .n-chip {
    align-self: flex-start; font-size: var(--app-text-xs); padding: 2px 8px;
    border-radius: var(--app-radius-pill); border: 1px solid var(--app-border-strong); color: var(--app-text-muted);
  }
  .n-chip.accent { background: var(--app-accent); color: #fff; border-color: transparent; }
  .n-flow { display: flex; align-items: center; gap: 4px; }
  .n-flow-step {
    flex: 1; min-width: 0; min-height: 1.4em;
    display: flex; align-items: center; justify-content: center;
    padding: 2px 4px;
    border-radius: var(--app-radius-sm); background: var(--app-accent-soft); border: 1px solid var(--app-border);
    color: var(--app-paper-ink); text-align: center; line-height: 1.1;
    overflow: hidden; white-space: nowrap; text-overflow: ellipsis;
  }
  .n-flow em { color: var(--app-text-subtle); font-style: normal; flex: 0 0 auto; }
  .n-cols { display: flex; gap: 6%; flex: 1; }
  .n-cols > div:empty { flex: 1; border: 1px solid var(--app-border); border-radius: var(--app-radius-sm); background: var(--app-surface-raised); }
  .n-col { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4%; }
  .n-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 5%; flex: 1; }
  .n-grid > div:empty { border: 1px solid var(--app-border); border-radius: var(--app-radius-sm); background: var(--app-surface-raised); }
  .n-cell { min-width: 0; display: flex; flex-direction: column; gap: 4%; }
  .n-card {
    display: flex; flex-direction: column; gap: 3%;
    border: 1px solid var(--app-border); border-radius: var(--app-radius-sm);
    background: var(--app-surface); padding: 5%;
    font-size: var(--app-text-xs); color: var(--app-text-muted);
    overflow: hidden;
  }
  .n-card-eyebrow {
    font-size: var(--app-text-xs);
    letter-spacing: 0.06em; text-transform: uppercase;
    color: var(--app-accent-text); opacity: 0.85;
    overflow: hidden; white-space: nowrap; text-overflow: ellipsis;
  }
  .n-card-header {
    font-weight: var(--app-weight-medium); color: var(--app-paper-ink);
    line-height: 1.15;
    overflow: hidden; white-space: nowrap; text-overflow: ellipsis;
  }

  .thumb.strip .paper { font-size: 5px; }
  .thumb.strip .hero-title { font-size: 9px; }
  .thumb.strip .n-heading { font-size: 7px; }
  .thumb.featured .paper { font-size: 14px; }
  .thumb.featured .hero-title { font-size: 34px; }
  .thumb.featured .n-heading { font-size: 20px; }

  .num {
    font-size: var(--app-text-xs);
    color: var(--app-text-muted);
    line-height: 1;
  }
  .thumb.strip.selected .num { color: var(--app-accent-text); font-weight: var(--app-weight-medium); }
</style>
