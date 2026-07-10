<script setup lang="ts">
import type { AcademicActivityItem } from "../../types/activity"
import AcademicActivityItemRow from "./AcademicActivityItem.vue"
import type { ActivityRole } from "./activityView"

defineProps<{
  groups: Array<{
    label: string
    items: AcademicActivityItem[]
  }>
  role: ActivityRole
}>()
</script>

<template>
  <div class="space-y-5">
    <section
      v-for="group in groups"
      :key="group.label"
      class="min-w-0"
      :aria-labelledby="`activity-group-${group.label}`"
    >
      <div class="mb-3 flex items-center justify-between gap-3">
        <h2
          :id="`activity-group-${group.label}`"
          class="text-sm font-medium text-foreground"
        >
          {{ group.label }}
        </h2>
        <span class="text-xs text-[#9ca3af]">{{ group.items.length }} item</span>
      </div>

      <ul class="space-y-2" aria-label="Aktivitas akademik">
        <li v-for="activity in group.items" :key="activity.id" class="min-w-0">
          <AcademicActivityItemRow :activity="activity" :role="role" />
        </li>
      </ul>
    </section>
  </div>
</template>
