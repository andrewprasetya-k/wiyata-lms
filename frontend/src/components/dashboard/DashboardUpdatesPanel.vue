<script setup lang="ts">
import { computed, ref } from "vue";

type UpdateTab = "notifications" | "chat" | "feed";

const props = defineProps<{
  notificationBadge?: number;
  chatBadge?: number;
  feedBadge?: number;
}>();

const activeTab = ref<UpdateTab>("notifications");

const tabs = computed(() => [
  {
    key: "notifications" as const,
    label: "Notifications",
    badge: props.notificationBadge ?? 0,
  },
  { key: "chat" as const, label: "Chat", badge: props.chatBadge ?? 0 },
  { key: "feed" as const, label: "Feed", badge: props.feedBadge ?? 0 },
]);

function badgeLabel(value: number) {
  if (value <= 0) return "";
  return value > 99 ? "99+" : String(value);
}

function selectTab(tab: UpdateTab) {
  activeTab.value = tab;
}

function handleTabKeydown(event: KeyboardEvent, index: number) {
  if (!["ArrowLeft", "ArrowRight", "Home", "End"].includes(event.key)) {
    return;
  }

  event.preventDefault();
  let nextIndex = index;
  if (event.key === "ArrowLeft") {
    nextIndex = index === 0 ? tabs.value.length - 1 : index - 1;
  }
  if (event.key === "ArrowRight") {
    nextIndex = index === tabs.value.length - 1 ? 0 : index + 1;
  }
  if (event.key === "Home") nextIndex = 0;
  if (event.key === "End") nextIndex = tabs.value.length - 1;

  activeTab.value = tabs.value[nextIndex].key;
}
</script>

<template>
  <section
    class="flex min-h-0 flex-col bg-surface"
    aria-labelledby="updates-panel-title"
  >
    <div class="shrink-0 px-2">
      <!-- <h2 id="updates-panel-title" class="text-sm font-medium text-foreground">
        Updates
      </h2> -->
      <div
        class="mt-3 flex gap-1 overflow-x-auto"
        role="tablist"
        aria-label="Updates"
      >
        <button
          v-for="(tab, index) in tabs"
          :id="`updates-tab-${tab.key}`"
          :key="tab.key"
          type="button"
          role="tab"
          class="inline-flex shrink-0 items-center gap-1.5 border-b-2 px-2 py-2 text-xs font-medium transition focus:outline-none focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-offset-2"
          :class="
            activeTab === tab.key
              ? 'border-brand text-brand'
              : 'border-transparent text-muted hover:text-foreground'
          "
          :aria-selected="activeTab === tab.key"
          :aria-controls="`updates-panel-${tab.key}`"
          :tabindex="activeTab === tab.key ? 0 : -1"
          @click="selectTab(tab.key)"
          @keydown="handleTabKeydown($event, index)"
        >
          {{ tab.label }}
          <span
            v-if="tab.badge > 0"
            class="inline-flex min-w-[1.1rem] items-center justify-center rounded-full bg-brand-soft px-1.5 py-0.5 text-[10px] font-semibold leading-none text-brand"
            :aria-label="`${tab.badge} update belum dibaca`"
          >
            {{ badgeLabel(tab.badge) }}
          </span>
        </button>
      </div>
    </div>

    <div class="min-h-0 flex-1 overflow-hidden p-3">
      <div
        v-show="activeTab === 'notifications'"
        id="updates-panel-notifications"
        role="tabpanel"
        aria-labelledby="updates-tab-notifications"
        class="h-full min-h-0 overflow-y-auto pr-1"
      >
        <slot name="notifications" />
      </div>
      <div
        v-show="activeTab === 'chat'"
        id="updates-panel-chat"
        role="tabpanel"
        aria-labelledby="updates-tab-chat"
        class="h-full min-h-0 overflow-y-auto pr-1"
      >
        <slot name="chat" />
      </div>
      <div
        v-show="activeTab === 'feed'"
        id="updates-panel-feed"
        role="tabpanel"
        aria-labelledby="updates-tab-feed"
        class="h-full min-h-0 overflow-y-auto pr-1"
      >
        <slot name="feed" />
      </div>
    </div>
  </section>
</template>
