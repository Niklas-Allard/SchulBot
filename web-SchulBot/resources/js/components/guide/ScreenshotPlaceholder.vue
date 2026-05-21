<script setup lang="ts">
defineProps<{ label: string }>();
</script>

<template>
    <div class="my-3 flex min-h-[140px] flex-col items-center justify-center gap-2 rounded-xl border-2 border-dashed border-muted-foreground/30 bg-muted/20 p-4 text-center">
        <div class="rounded-full bg-muted p-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
        </div>
        <p class="max-w-sm text-xs text-muted-foreground">📸 {{ label }}</p>
    </div>
</template>
