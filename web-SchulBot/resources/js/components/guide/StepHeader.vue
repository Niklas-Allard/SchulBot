<script setup lang="ts">
defineProps<{
    step: string;
    title: string;
    badge?: string;
    badgeVariant?: 'default' | 'secondary' | 'outline' | 'destructive';
}>();
</script>

<template>
    <div class="mb-6 flex items-start gap-3">
        <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary text-sm font-bold text-primary-foreground">
            {{ step }}
        </div>
        <div>
            <h2 class="text-xl font-semibold">{{ title }}</h2>
            <span
                v-if="badge"
                class="mt-1 inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold"
                :class="
                    badgeVariant === 'default'   ? 'border-transparent bg-primary text-primary-foreground' :
                    badgeVariant === 'secondary' ? 'border-transparent bg-secondary text-secondary-foreground' :
                                                   'border-border text-muted-foreground'
                "
            >
                {{ badge }}
            </span>
        </div>
    </div>
</template>