<script setup lang="ts">
import { Head, Link } from '@inertiajs/vue3';
import { ref } from 'vue';
import Heading from '@/components/Heading.vue';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';

interface HistoryEntry {
    id: number;
    tag: string;
    payload: string;
    response: string | null;
    sender_email: string;
    status: 'ok' | 'error';
    created_at: string;
}

interface Paginator {
    data: HistoryEntry[];
    current_page: number;
    last_page: number;
    next_page_url: string | null;
    prev_page_url: string | null;
    from: number;
    to: number;
    total: number;
}

defineProps<{
    history: Paginator;
}>();

const expanded = ref<Set<number>>(new Set());

function toggle(id: number) {
    if (expanded.value.has(id)) {
        expanded.value.delete(id);
    } else {
        expanded.value.add(id);
    }
}

function formatDate(iso: string): string {
    return new Date(iso).toLocaleString('de-DE', { dateStyle: 'short', timeStyle: 'short' });
}

defineOptions({
    layout: {
        breadcrumbs: [
            { title: 'Dashboard', href: '/dashboard' },
            { title: 'Verlauf', href: '/bot/history' },
        ],
    },
});
</script>

<template>
    <Head title="Befehlsverlauf" />

    <div class="flex flex-col gap-6 p-4">
        <Heading title="Befehlsverlauf" description="Alle von deinem Bot verarbeiteten Befehle." />

        <div v-if="history.data.length === 0" class="rounded-xl border border-dashed p-8 text-center text-sm text-muted-foreground">
            Noch keine Befehle verarbeitet. Schreibe eine E-Mail mit <code class="font-mono">#ki</code> an deinen Bot!
        </div>

        <div v-else class="overflow-hidden rounded-xl border">
            <table class="w-full text-sm">
                <thead class="border-b bg-muted/40">
                    <tr>
                        <th class="px-4 py-3 text-left font-medium">Tag</th>
                        <th class="px-4 py-3 text-left font-medium">Payload</th>
                        <th class="px-4 py-3 text-left font-medium">Status</th>
                        <th class="px-4 py-3 text-left font-medium">Datum</th>
                        <th class="px-4 py-3 text-left font-medium"></th>
                    </tr>
                </thead>
                <tbody>
                    <template v-for="entry in history.data" :key="entry.id">
                        <tr class="border-b transition-colors hover:bg-muted/30">
                            <td class="px-4 py-3">
                                <code class="rounded bg-muted px-1 py-0.5 font-mono text-xs">{{ entry.tag }}</code>
                            </td>
                            <td class="max-w-xs px-4 py-3">
                                <span class="line-clamp-1 text-muted-foreground">{{ entry.payload }}</span>
                            </td>
                            <td class="px-4 py-3">
                                <Badge :variant="entry.status === 'ok' ? 'default' : 'destructive'">
                                    {{ entry.status === 'ok' ? 'OK' : 'Fehler' }}
                                </Badge>
                            </td>
                            <td class="whitespace-nowrap px-4 py-3 text-muted-foreground">
                                {{ formatDate(entry.created_at) }}
                            </td>
                            <td class="px-4 py-3">
                                <Button
                                    v-if="entry.response"
                                    variant="ghost"
                                    size="sm"
                                    @click="toggle(entry.id)"
                                >
                                    {{ expanded.has(entry.id) ? 'Einklappen' : 'Antwort' }}
                                </Button>
                            </td>
                        </tr>
                        <tr v-if="expanded.has(entry.id)" class="border-b bg-muted/20">
                            <td colspan="5" class="px-4 py-3">
                                <pre class="whitespace-pre-wrap font-mono text-xs">{{ entry.response }}</pre>
                            </td>
                        </tr>
                    </template>
                </tbody>
            </table>
        </div>

        <!-- Pagination -->
        <div v-if="history.last_page > 1" class="flex items-center justify-between text-sm text-muted-foreground">
            <span>{{ history.from }}–{{ history.to }} von {{ history.total }}</span>
            <div class="flex gap-2">
                <Button
                    variant="outline"
                    size="sm"
                    as-child
                    :disabled="!history.prev_page_url"
                >
                    <Link v-if="history.prev_page_url" :href="history.prev_page_url">Zurück</Link>
                    <span v-else>Zurück</span>
                </Button>
                <Button
                    variant="outline"
                    size="sm"
                    as-child
                    :disabled="!history.next_page_url"
                >
                    <Link v-if="history.next_page_url" :href="history.next_page_url">Weiter</Link>
                    <span v-else>Weiter</span>
                </Button>
            </div>
        </div>
    </div>
</template>
