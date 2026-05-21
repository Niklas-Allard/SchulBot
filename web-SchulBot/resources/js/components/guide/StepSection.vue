<script setup lang="ts">
defineProps<{ number: string; title: string }>();
</script>

<template>
    <div class="flex gap-4">
        <div class="flex flex-col items-center gap-1">
            <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-semibold uppercase text-muted-foreground ring-1 ring-border">
                {{ number }}
            </span>
            <div class="w-px flex-1 bg-border" />
        </div>
        <div class="flex-1 pb-6">
            <h3 class="mb-2 font-semibold">{{ title }}</h3>
            <slot />
        </div>
    </div>
</template>
