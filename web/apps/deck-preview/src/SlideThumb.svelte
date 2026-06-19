<!--
  SlideThumb — renders one SlidePreview as a miniature slide, natively from the
  IR node descriptors (no server-side raster). The thumbnail sits on the deck
  "paper" (--app-paper) and approximates each node kind with a styled block, so
  the strip reads like the real deck at a glance. Two sizes: 'featured' (the
  large hero thumbnail) and 'strip' (the small filmstrip items).
-->
<script lang="ts">
  export interface ThumbNode {
    kind: string;
    text?: string;
    detail?: string;
    count?: number;
    accent?: boolean;
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
</script>

<button
  type="button"
  class="thumb {size}"
  class:selected
  aria-label={`Slide ${slide.index + 1}${slide.title ? `: ${slide.title}` : ''}`}
  onclick={() => onselect?.(slide.index)}
>
  <div class="paper" class:cover={isCover}>
    {#each nodes as node (node.kind + (node.text ?? '') + node.count)}
      {#if node.kind === 'hero'}
        <div class="n-hero">
          {#if node.detail}<span class="eyebrow">{node.detail}</span>{/if}
          <span class="hero-title">{node.text}</span>
        </div>
      {:else if node.kind === 'heading' || node.kind === 'section_divider'}
        <span class="n-heading">{node.text}</span>
      {:else if node.kind === 'prose'}
        <span class="n-line"></span><span class="n-line short"></span>
      {:else if node.kind === 'list'}
        {#each Array(Math.min(node.count || 3, 4)) as _, i (i)}
          <span class="n-bullet"><i></i><b></b></span>
        {/each}
      {:else if node.kind === 'callout'}
        <div class="n-callout" class:accent={node.accent}>{node.text || 'Callout'}</div>
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
          {#each Array(Math.min(node.count || 3, 4)) as _, i (i)}
            <span></span>{#if i < Math.min(node.count || 3, 4) - 1}<em>›</em>{/if}
          {/each}
        </div>
      {:else if node.kind === 'two_column'}
        <div class="n-cols"><div></div><div></div></div>
      {:else if node.kind === 'grid'}
        <div class="n-grid">
          {#each Array(Math.min(node.count || 4, 6)) as _, i (i)}<div></div>{/each}
        </div>
      {:else if node.kind === 'card' || node.kind === 'card_section'}
        <div class="n-card">{node.text || ''}</div>
      {:else}
        <span class="n-line"></span>
      {/if}
    {/each}
    {#if nodes.length === 0}
      <span class="n-heading muted">{slide.title || 'Empty slide'}</span>
    {/if}
  </div>
  {#if size === 'strip'}
    <span class="num">{slide.index + 1}</span>
  {/if}
</button>

<style>
  .thumb {
    display: block;
    padding: 0;
    border: 1px solid var(--app-border);
    border-radius: var(--app-radius-md);
    background: var(--app-surface);
    cursor: pointer;
    overflow: hidden;
    position: relative;
    transition: border-color var(--app-dur) var(--app-ease),
      box-shadow var(--app-dur) var(--app-ease), transform var(--app-dur) var(--app-ease);
  }
  .thumb:hover { border-color: var(--app-accent); box-shadow: var(--app-shadow-sm); }
  .thumb:focus-visible { outline: 2px solid var(--app-accent); outline-offset: 2px; }
  .thumb.selected { border-color: var(--app-accent); box-shadow: var(--app-shadow-md); }

  .thumb.featured { width: 100%; }
  .thumb.strip { width: 96px; flex: 0 0 auto; }

  /* 16:9 paper */
  .paper {
    aspect-ratio: 16 / 9;
    background: var(--app-paper);
    color: var(--app-paper-ink);
    display: flex;
    flex-direction: column;
    gap: 3%;
    padding: 7% 8%;
    overflow: hidden;
  }
  .paper.cover { justify-content: center; }

  .thumb.featured .paper { gap: 2.4%; padding: 6% 7%; }

  /* node blocks — scale with the paper via cqw-ish relative units */
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
  .n-line {
    height: 0.5em;
    border-radius: 2px;
    background: color-mix(in srgb, var(--app-paper-ink) 22%, transparent);
  }
  .n-line.short { width: 62%; }
  .n-bullet { display: flex; align-items: center; gap: 5px; }
  .n-bullet i { width: 5px; height: 5px; border-radius: 50%; background: var(--app-accent); flex: 0 0 auto; }
  .n-bullet b { height: 0.42em; flex: 1; border-radius: 2px; background: color-mix(in srgb, var(--app-paper-ink) 20%, transparent); }
  .n-callout {
    border-left: 3px solid var(--app-accent);
    background: var(--app-accent-soft);
    border-radius: var(--app-radius-sm);
    padding: 5% 6%;
    font-size: var(--app-text-xs);
    color: var(--app-paper-ink);
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }
  .n-quote {
    font-family: var(--app-font-serif);
    font-style: italic;
    border-left: 3px solid var(--app-border-strong);
    padding-left: 6%;
    color: var(--app-text-muted);
    overflow: hidden;
  }
  .n-chart { display: flex; align-items: flex-end; gap: 8%; height: 38%; }
  .n-chart i { flex: 1; background: var(--app-accent); border-radius: 2px 2px 0 0; opacity: 0.9; }
  .n-code {
    background: #2b2723;
    border-radius: var(--app-radius-sm);
    padding: 6%;
    display: flex; flex-direction: column; gap: 5px;
    flex: 1;
  }
  .n-code i { height: 0.4em; border-radius: 2px; background: rgba(250, 247, 242, 0.55); width: 80%; }
  .n-code i.ind { width: 60%; margin-left: 12%; }
  .n-table { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 3px; }
  .n-table span { height: 0.7em; background: color-mix(in srgb, var(--app-paper-ink) 12%, transparent); border-radius: 2px; }
  .n-table span:nth-child(-n+3) { background: var(--app-accent-soft); }
  .n-image { flex: 1; min-height: 30%; background: color-mix(in srgb, var(--app-paper-ink) 10%, transparent); border-radius: var(--app-radius-sm); }
  .n-divider { height: 1px; background: var(--app-border-strong); }
  .n-chip {
    align-self: flex-start;
    font-size: var(--app-text-xs);
    padding: 2px 8px;
    border-radius: var(--app-radius-pill);
    border: 1px solid var(--app-border-strong);
    color: var(--app-text-muted);
  }
  .n-chip.accent { background: var(--app-accent); color: #fff; border-color: transparent; }
  .n-flow { display: flex; align-items: center; gap: 4px; }
  .n-flow span { flex: 1; height: 1.4em; border-radius: var(--app-radius-sm); background: var(--app-accent-soft); border: 1px solid var(--app-border); }
  .n-flow em { color: var(--app-text-subtle); font-style: normal; }
  .n-cols { display: flex; gap: 6%; flex: 1; }
  .n-cols div { flex: 1; border: 1px solid var(--app-border); border-radius: var(--app-radius-sm); background: var(--app-surface-raised); }
  .n-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 5%; flex: 1; }
  .n-grid div { border: 1px solid var(--app-border); border-radius: var(--app-radius-sm); background: var(--app-surface-raised); }
  .n-card { border: 1px solid var(--app-border); border-radius: var(--app-radius-sm); background: var(--app-surface); padding: 4%; font-size: var(--app-text-xs); color: var(--app-text-muted); }

  /* font scale per size */
  .thumb.strip .paper { font-size: 5px; }
  .thumb.strip .hero-title { font-size: 9px; }
  .thumb.strip .n-heading { font-size: 7px; }
  .thumb.featured .paper { font-size: 13px; }
  .thumb.featured .hero-title { font-size: 30px; }
  .thumb.featured .n-heading { font-size: 19px; }

  .num {
    position: absolute;
    bottom: 4px; right: 5px;
    font-size: var(--app-text-xs);
    color: var(--app-text-muted);
    background: var(--app-surface);
    border-radius: var(--app-radius-sm);
    padding: 0 4px;
    line-height: 1.4;
  }
</style>
