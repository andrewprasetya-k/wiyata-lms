import { ref } from "vue";
import { defineStore } from "pinia";

export type ConfirmVariant = "danger" | "warning" | "default";

export interface ConfirmOptions {
  title: string;
  description?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  variant?: ConfirmVariant;
}

export const useConfirmStore = defineStore("confirm", () => {
  const open = ref(false);
  const options = ref<ConfirmOptions>({ title: "" });

  let resolver: ((value: boolean) => void) | null = null;

  function confirm(opts: ConfirmOptions): Promise<boolean> {
    options.value = opts;
    open.value = true;
    return new Promise<boolean>((resolve) => {
      resolver = resolve;
    });
  }

  function _settle(value: boolean) {
    open.value = false;
    const r = resolver;
    resolver = null;
    r?.(value);
  }

  function accept() {
    _settle(true);
  }

  function dismiss() {
    _settle(false);
  }

  return { open, options, confirm, accept, dismiss };
});
