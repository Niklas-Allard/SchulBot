<script setup lang="ts">
import { Head, Link, usePage } from '@inertiajs/vue3';
import { computed } from 'vue';
import Heading from '@/components/Heading.vue';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { dashboard } from '@/routes';

interface RecentEntry {
    id: number;
    tag: string;
    payload: string;
    status: 'ok' | 'error';
    created_at: string;
}

const props = defineProps<{
    botConfigured: boolean;
    botActive: boolean;
    commandsTotal: number;
    commandsToday: number;
    recent: RecentEntry[];
}>();

const page = usePage();
const user = computed(() => page.props.auth.user as any);

function formatDate(iso: string): string {
    return new Date(iso).toLocaleString('de-DE', { dateStyle: 'short', timeStyle: 'short' });
}

defineOptions({
    layout: {
        breadcrumbs: [
            { title: 'Dashboard', href: dashboard() },
        ],
    },
});
</script>

<template>
    <Head title="Dashboard" />

    <div class="flex flex-col gap-6 p-4">
        <Heading :title="`Hallo, ${user.name}!`" description="Dein SchulBot-Übersicht" />

        <!-- Bot not set up banner -->
        <div
            v-if="!botConfigured"
            class="flex items-center justify-between rounded-xl border border-yellow-200 bg-yellow-50 px-4 py-3 dark:border-yellow-800 dark:bg-yellow-950"
        >
            <p class="text-sm text-yellow-800 dark:text-yellow-200">
                Dein Bot ist noch nicht eingerichtet. Gib deine IServ-Zugangsdaten ein, um loszulegen.
            </p>
            <Button as-child variant="outline" size="sm" class="ml-4 shrink-0">
                <Link href="/bot/setup">Bot einrichten</Link>
            </Button>
        </div>

        <!-- Stats -->
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <Card>
                <CardHeader class="pb-2">
                    <CardTitle class="text-sm font-medium text-muted-foreground">Bot-Status</CardTitle>
                </CardHeader>
                <CardContent>
                    <div class="flex items-center gap-2">
                        <div
                            class="h-2 w-2 rounded-full"
                            :class="botActive ? 'bg-green-500' : 'bg-gray-400'"
                        />
                        <span class="text-lg font-semibold">{{ botActive ? 'Aktiv' : 'Inaktiv' }}</span>
                    </div>
                </CardContent>
            </Card>
            <Card>
                <CardHeader class="pb-2">
                    <CardTitle class="text-sm font-medium text-muted-foreground">Befehle gesamt</CardTitle>
                </CardHeader>
                <CardContent>
                    <p class="text-2xl font-bold">{{ commandsTotal }}</p>
                </CardContent>
            </Card>
            <Card>
                <CardHeader class="pb-2">
                    <CardTitle class="text-sm font-medium text-muted-foreground">Befehle heute</CardTitle>
                </CardHeader>
                <CardContent>
                    <p class="text-2xl font-bold">{{ commandsToday }}</p>
                </CardContent>
            </Card>
            <Card>
                <CardHeader class="pb-2">
                    <CardTitle class="text-sm font-medium text-muted-foreground">Verfügbare Befehle</CardTitle>
                </CardHeader>
                <CardContent>
                    <p class="text-2xl font-bold">8</p>
                </CardContent>
            </Card>
        </div>

        <!-- Recent activity -->
        <Card>
            <CardHeader class="flex flex-row items-center justify-between">
                <div>
                    <CardTitle>Letzte Aktivität</CardTitle>
                    <CardDescription>Deine letzten 5 verarbeiteten Befehle</CardDescription>
                </div>
                <Button as-child variant="outline" size="sm">
                    <Link href="/bot/history">Alle anzeigen</Link>
                </Button>
            </CardHeader>
            <CardContent>
                <div v-if="recent.length === 0" class="py-6 text-center text-sm text-muted-foreground">
                    Noch keine Befehle. Schreib eine E-Mail mit <code class="font-mono">#ki Deine Frage</code> an deinen Bot!
                </div>
                <table v-else class="w-full text-sm">
                    <tbody>
                        <tr
                            v-for="entry in recent"
                            :key="entry.id"
                            class="border-b last:border-0"
                        >
                            <td class="py-2 pr-4">
                                <code class="rounded bg-muted px-1 py-0.5 font-mono text-xs">{{ entry.tag }}</code>
                            </td>
                            <td class="max-w-xs py-2 pr-4">
                                <span class="line-clamp-1 text-muted-foreground">{{ entry.payload }}</span>
                            </td>
                            <td class="py-2 pr-4">
                                <Badge :variant="entry.status === 'ok' ? 'default' : 'destructive'" class="text-xs">
                                    {{ entry.status === 'ok' ? 'OK' : 'Fehler' }}
                                </Badge>
                            </td>
                            <td class="whitespace-nowrap py-2 text-xs text-muted-foreground">
                                {{ formatDate(entry.created_at) }}
                            </td>
                        </tr>
                    </tbody>
                </table>
            </CardContent>
        </Card>

        <!-- Commands reference -->
        <Card>
            <CardHeader>
                <CardTitle>Verfügbare Befehle</CardTitle>
                <CardDescription>Sende eine E-Mail mit einem dieser Tags als Betreff</CardDescription>
            </CardHeader>
            <CardContent>
                <div class="grid gap-2 sm:grid-cols-2">
                    <div v-for="cmd in [
                        { tag: '#ki', desc: 'KI-Frage stellen' },
                        { tag: '#news', desc: 'Aktuelle Nachrichten' },
                        { tag: '#sudoku', desc: 'Sudoku-Rätsel' },
                        { tag: '#translate', desc: 'Text übersetzen' },
                        { tag: '#tasks', desc: 'Google Tasks' },
                        { tag: '#calendar', desc: 'Google Kalender' },
                        { tag: '#model', desc: 'KI-Modell wechseln' },
                        { tag: '#hilfe', desc: 'Hilfe anzeigen' },
                    ]" :key="cmd.tag" class="flex items-center gap-3 rounded-lg border p-3">
                        <code class="font-mono text-sm font-medium">{{ cmd.tag }}</code>
                        <span class="text-sm text-muted-foreground">{{ cmd.desc }}</span>
                    </div>
                </div>
            </CardContent>
        </Card>
    </div>
</template>
