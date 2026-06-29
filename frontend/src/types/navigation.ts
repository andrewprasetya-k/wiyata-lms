import type { Component } from 'vue'

export interface NavItem {
  label: string
  to: string
  icon: Component
  hasDot?: boolean
  badgeCount?: number
  badgeLabel?: string
  badgeAriaLabel?: string
  emphasized?: boolean
}
