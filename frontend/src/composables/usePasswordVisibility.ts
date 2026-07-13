import { computed, ref } from "vue";

export function usePasswordVisibility(initialVisible = false) {
  const visible = ref(initialVisible);
  const inputType = computed(() => (visible.value ? "text" : "password"));

  function toggle() {
    visible.value = !visible.value;
  }

  return { visible, inputType, toggle };
}