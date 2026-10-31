<!-- One status dot, drawn from the tone of the rule in status.ts. The text of
     the state lives in the row or the line around it, so the dot itself is
     hidden from screen readers. -->
<script lang="ts">
  import type { Tone } from './status'

  let { tone }: { tone: Tone } = $props()
</script>

<span class="dot {tone}" aria-hidden="true"></span>

<style>
  .dot {
    flex: none;
    box-sizing: border-box;
    width: var(--dot-size);
    height: var(--dot-size);
    border-radius: var(--dot-size);
  }
  .dot.positive {
    background: var(--positive);
  }
  .dot.warning {
    background: var(--warning);
  }
  .dot.negative {
    background: var(--negative);
  }
  .dot.muted {
    background: var(--muted);
  }
  /* Rimosso: a ring, so it never reads as one of the filled states. */
  .dot.outline {
    border: 1px solid var(--control);
    background: transparent;
  }
</style>
