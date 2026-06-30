<script setup lang="ts">
import { RouterLink } from "vue-router"
import { PhArrowRight } from "@phosphor-icons/vue"
import type { AcademicActivityItem } from "../../types/activity"
import ActivityTypeBadge from "./ActivityTypeBadge.vue"
import {
  activityRelativeLabel,
  activitySubjectColor,
  activityTypeLabel,
  isInternalActivityLink,
  type ActivityRole,
} from "./activityView"

const props = defineProps<{
  activity: AcademicActivityItem
  role: ActivityRole
}>()

function ariaLabel() {
  return `${activityTypeLabel(props.activity.type, props.role)}: ${
    props.activity.title
  }`
}
</script>

<template>
  <RouterLink
    v-if="isInternalActivityLink(activity.link)"
    :to="activity.link || ''"
    class="group flex min-w-0 gap-3 rounded-lg border border-[#ebe7df] bg-white p-4 transition hover:border-[#c7d2fe] hover:bg-[#fbfaf8] focus:outline-none focus-visible:ring-2 focus-visible:ring-[#4f46e5] focus-visible:ring-offset-2"
    :aria-label="ariaLabel()"
  >
    <span
      class="mt-2 h-2.5 w-2.5 shrink-0 rounded-full"
      :style="{ backgroundColor: activitySubjectColor(activity) }"
      aria-hidden="true"
    />
    <span class="min-w-0 flex-1">
      <span class="flex min-w-0 flex-wrap items-center gap-2">
        <ActivityTypeBadge :label="activityTypeLabel(activity.type, role)" />
        <span class="text-[11px] text-[#8b8592]">
          {{ activityRelativeLabel(activity) }}
        </span>
        <span
          v-if="activity.priority === 'high'"
          class="rounded-full bg-[#fff7ed] px-2 py-0.5 text-[10px] font-medium text-[#b45309]"
        >
          Prioritas
        </span>
      </span>
      <span
        class="mt-2 line-clamp-2 text-sm font-medium leading-5 text-[#171322] transition group-hover:text-[#4f46e5]"
      >
        {{ activity.title }}
      </span>
      <span class="mt-1 line-clamp-2 text-xs leading-5 text-[#7a7385]">
        {{ activity.description }}
      </span>
      <span
        v-if="activity.class?.name || activity.subject?.name"
        class="mt-2 block truncate text-[11px] text-[#9ca3af]"
      >
        {{ activity.subject?.name || "Mata pelajaran" }}
        <template v-if="activity.class?.name"> · {{ activity.class.name }}</template>
      </span>
    </span>
    <PhArrowRight
      :size="15"
      class="mt-2 shrink-0 text-[#a09aa8] transition group-hover:translate-x-0.5 group-hover:text-[#4f46e5]"
      aria-hidden="true"
    />
  </RouterLink>

  <article
    v-else
    class="flex min-w-0 gap-3 rounded-lg border border-[#ebe7df] bg-white p-4"
  >
    <span
      class="mt-2 h-2.5 w-2.5 shrink-0 rounded-full"
      :style="{ backgroundColor: activitySubjectColor(activity) }"
      aria-hidden="true"
    />
    <div class="min-w-0 flex-1">
      <div class="flex min-w-0 flex-wrap items-center gap-2">
        <ActivityTypeBadge :label="activityTypeLabel(activity.type, role)" />
        <span class="text-[11px] text-[#8b8592]">
          {{ activityRelativeLabel(activity) }}
        </span>
        <span
          v-if="activity.priority === 'high'"
          class="rounded-full bg-[#fff7ed] px-2 py-0.5 text-[10px] font-medium text-[#b45309]"
        >
          Prioritas
        </span>
      </div>
      <h3 class="mt-2 line-clamp-2 text-sm font-medium text-[#171322]">
        {{ activity.title }}
      </h3>
      <p class="mt-1 line-clamp-2 text-xs leading-5 text-[#7a7385]">
        {{ activity.description }}
      </p>
      <p
        v-if="activity.class?.name || activity.subject?.name"
        class="mt-2 truncate text-[11px] text-[#9ca3af]"
      >
        {{ activity.subject?.name || "Mata pelajaran" }}
        <template v-if="activity.class?.name"> · {{ activity.class.name }}</template>
      </p>
    </div>
  </article>
</template>
