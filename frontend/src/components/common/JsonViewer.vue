<script setup lang="ts">
defineProps<{ value: unknown }>();

function isObject(v: unknown): v is Record<string, unknown> {
  return typeof v === "object" && v !== null && !Array.isArray(v);
}

function isArrayValue(v: unknown): v is unknown[] {
  return Array.isArray(v);
}

function isExpandable(v: unknown): boolean {
  return isObject(v) || isArrayValue(v);
}

function displayValue(v: unknown): string {
  if (v === null) return "null";
  if (v === undefined) return "—";
  if (typeof v === "boolean") return v ? "true" : "false";
  if (typeof v === "string" && v.trim() === "") return "—";
  return String(v);
}
</script>

<template>
  <p
    v-if="!isExpandable(value)"
    class="break-all font-mono text-xs text-foreground"
  >
    {{ displayValue(value) }}
  </p>
  <p
    v-else-if="isObject(value) && Object.keys(value).length === 0"
    class="text-xs text-muted"
  >
    Tidak ada data.
  </p>
  <ul v-else class="space-y-2">
    <li
      v-for="(entryValue, key) in value as Record<string | number, unknown>"
      :key="key"
      class="text-sm"
    >
      <div class="flex flex-wrap items-baseline gap-x-2 gap-y-1">
        <span
          class="shrink-0 rounded bg-surface-subtle px-1.5 py-0.5 font-mono text-[11px] font-medium text-muted"
        >
          {{ key }}
        </span>
        <span
          v-if="!isExpandable(entryValue)"
          class="break-all font-mono text-xs text-foreground"
        >
          {{ displayValue(entryValue) }}
        </span>
      </div>
      <div
        v-if="isExpandable(entryValue)"
        class="mt-1.5 border-l border-border pl-3"
      >
        <JsonViewer :value="entryValue" />
      </div>
    </li>
  </ul>
</template>
