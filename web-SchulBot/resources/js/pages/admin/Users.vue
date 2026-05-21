<script setup lang="ts">
import { Head, router } from '@inertiajs/vue3';
import Heading from '@/components/Heading.vue';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

interface UserRow {
    id: number;
    name: string;
    email: string;
    role: string;
    bot_active: boolean;
    bot_set_up: boolean;
    created_at: string;
}

interface Stats {
    total_users: number;
    active_bots: number;
    commands_today: number;
    commands_total: number;
}

defineProps<{
    users: UserRow[];
    stats: Stats;
}>();

function toggleBot(userId: number) {
    router.post(`/admin/users/${userId}/toggle-bot`);
}

function makeAdmin(userId: number) {
    if (confirm('Diesen Nutzer zum Admin machen?')) {
        router.post(`/admin/users/${userId}/make-admin`);
    }
}

function destroyUser(userId: number) {
    if (confirm('Nutzer wirklich löschen? Diese Aktion kann nicht rückgängig gemacht werden.')) {
        router.delete(`/admin/users/${userId}`);
    }
}

function formatDate(iso: string): string {
    return new Date(iso).toLocaleDateString('de-DE');
}

defineOptions({
    layout: {
        breadcrumbs: [
            { title: 'Dashboard', href: '/dashboard' },
            { title: 'Admin', href: '/admin/users' },
        ],
    },
});
</script>

<template>
    <Head title="Admin – Nutzer" />

    <div class="flex flex-col gap-6 p-4">
        <Heading title="Nutzerverwaltung" description="Übersicht aller registrierten Schulkameraden." />

        <!-- Stats -->
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <Card>
                <CardHeader class="pb-2"><CardTitle class="text-sm font-medium text-muted-foreground">Nutzer gesamt</CardTitle></CardHeader>
                <CardContent><p class="text-2xl font-bold">{{ stats.total_users }}</p></CardContent>
            </Card>
            <Card>
                <CardHeader class="pb-2"><CardTitle class="text-sm font-medium text-muted-foreground">Aktive Bots</CardTitle></CardHeader>
                <CardContent><p class="text-2xl font-bold">{{ stats.active_bots }}</p></CardContent>
            </Card>
            <Card>
                <CardHeader class="pb-2"><CardTitle class="text-sm font-medium text-muted-foreground">Befehle heute</CardTitle></CardHeader>
                <CardContent><p class="text-2xl font-bold">{{ stats.commands_today }}</p></CardContent>
            </Card>
            <Card>
                <CardHeader class="pb-2"><CardTitle class="text-sm font-medium text-muted-foreground">Befehle gesamt</CardTitle></CardHeader>
                <CardContent><p class="text-2xl font-bold">{{ stats.commands_total }}</p></CardContent>
            </Card>
        </div>

        <!-- Users table -->
        <div class="overflow-hidden rounded-xl border">
            <table class="w-full text-sm">
                <thead class="border-b bg-muted/40">
                    <tr>
                        <th class="px-4 py-3 text-left font-medium">Name</th>
                        <th class="px-4 py-3 text-left font-medium">E-Mail</th>
                        <th class="px-4 py-3 text-left font-medium">Rolle</th>
                        <th class="px-4 py-3 text-left font-medium">Bot</th>
                        <th class="px-4 py-3 text-left font-medium">Registriert</th>
                        <th class="px-4 py-3 text-left font-medium">Aktionen</th>
                    </tr>
                </thead>
                <tbody>
                    <tr
                        v-for="user in users"
                        :key="user.id"
                        class="border-b transition-colors hover:bg-muted/30"
                    >
                        <td class="px-4 py-3 font-medium">{{ user.name }}</td>
                        <td class="px-4 py-3 text-muted-foreground">{{ user.email }}</td>
                        <td class="px-4 py-3">
                            <Badge :variant="user.role === 'admin' ? 'default' : 'secondary'">
                                {{ user.role }}
                            </Badge>
                        </td>
                        <td class="px-4 py-3">
                            <Badge v-if="!user.bot_set_up" variant="outline" class="text-muted-foreground">
                                nicht eingerichtet
                            </Badge>
                            <Badge v-else-if="user.bot_active" variant="default">aktiv</Badge>
                            <Badge v-else variant="secondary">pausiert</Badge>
                        </td>
                        <td class="whitespace-nowrap px-4 py-3 text-muted-foreground">
                            {{ formatDate(user.created_at) }}
                        </td>
                        <td class="px-4 py-3">
                            <div class="flex gap-2">
                                <Button
                                    v-if="user.bot_set_up"
                                    variant="outline"
                                    size="sm"
                                    @click="toggleBot(user.id)"
                                >
                                    {{ user.bot_active ? 'Pausieren' : 'Aktivieren' }}
                                </Button>
                                <Button
                                    v-if="user.role !== 'admin'"
                                    variant="outline"
                                    size="sm"
                                    @click="makeAdmin(user.id)"
                                >
                                    Admin
                                </Button>
                                <Button
                                    variant="destructive"
                                    size="sm"
                                    @click="destroyUser(user.id)"
                                >
                                    Löschen
                                </Button>
                            </div>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>
