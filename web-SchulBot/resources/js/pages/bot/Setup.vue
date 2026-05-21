<script setup lang="ts">
import { Head, useForm, usePage } from '@inertiajs/vue3';
import Heading from '@/components/Heading.vue';
import InputError from '@/components/InputError.vue';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

interface BotConfigData {
    smtp_host: string;
    smtp_port: number;
    smtp_username: string;
    smtp_password: string;
    smtp_security: string;
    smtp_from_name: string;
    smtp_from_address: string;
    ai_provider: string;
    ai_api_url: string;
    ai_api_key: string;
    ai_model: string;
    is_active: boolean;
}

const props = defineProps<{
    config: Omit<BotConfigData, 'smtp_password' | 'ai_api_key'> | null;
}>();

const page = usePage();
const isEditing = props.config !== null;

const form = useForm<BotConfigData>({
    smtp_host:         props.config?.smtp_host         ?? 'pggv.de',
    smtp_port:         props.config?.smtp_port         ?? 465,
    smtp_username:     props.config?.smtp_username     ?? '',
    smtp_password:     '',
    smtp_security:     props.config?.smtp_security     ?? 'SSL',
    smtp_from_name:    props.config?.smtp_from_name    ?? (page.props.auth.user as any).name ?? '',
    smtp_from_address: props.config?.smtp_from_address ?? '',
    ai_provider:       props.config?.ai_provider       ?? 'gemini',
    ai_api_url:        props.config?.ai_api_url        ?? '',
    ai_api_key:        '',
    ai_model:          props.config?.ai_model          ?? '',
    is_active:         props.config?.is_active         ?? true,
});

function submit() {
    form.post('/bot/setup');
}

defineOptions({
    layout: {
        breadcrumbs: [
            { title: 'Dashboard', href: '/dashboard' },
            { title: 'Bot-Einrichtung', href: '/bot/setup' },
        ],
    },
});
</script>

<template>
    <Head title="Bot-Einrichtung" />

    <div class="flex flex-col gap-6 p-4">
        <Heading
            title="Bot-Einrichtung"
            description="E-Mails werden über den gemeinsamen Bot-Account empfangen. Konfiguriere hier deinen eigenen Absender für Antworten."
        />

        <form @submit.prevent="submit" class="flex flex-col gap-6">
            <!-- SMTP -->
            <Card>
                <CardHeader>
                    <CardTitle>E-Mail senden (SMTP)</CardTitle>
                </CardHeader>
                <CardContent class="grid gap-4 sm:grid-cols-2">
                    <div class="grid gap-2">
                        <Label for="smtp_host">Server</Label>
                        <Input id="smtp_host" v-model="form.smtp_host" placeholder="pggv.de" />
                        <InputError :message="form.errors.smtp_host" />
                    </div>
                    <div class="grid gap-2">
                        <Label for="smtp_port">Port</Label>
                        <Input id="smtp_port" type="number" v-model="form.smtp_port" />
                        <InputError :message="form.errors.smtp_port" />
                    </div>
                    <div class="grid gap-2">
                        <Label for="smtp_username">Benutzername</Label>
                        <Input id="smtp_username" v-model="form.smtp_username" placeholder="vorname.nachname" autocomplete="off" />
                        <InputError :message="form.errors.smtp_username" />
                    </div>
                    <div class="grid gap-2">
                        <Label for="smtp_password">Passwort{{ isEditing ? ' (leer = unverändert)' : '' }}</Label>
                        <Input id="smtp_password" type="password" v-model="form.smtp_password" autocomplete="new-password" />
                        <InputError :message="form.errors.smtp_password" />
                    </div>
                    <div class="grid gap-2">
                        <Label for="smtp_security">Verschlüsselung</Label>
                        <Select v-model="form.smtp_security">
                            <SelectTrigger id="smtp_security"><SelectValue /></SelectTrigger>
                            <SelectContent>
                                <SelectItem value="SSL">SSL/TLS (Port 465)</SelectItem>
                                <SelectItem value="STARTTLS">STARTTLS (Port 587)</SelectItem>
                            </SelectContent>
                        </Select>
                        <InputError :message="form.errors.smtp_security" />
                    </div>
                    <div class="grid gap-2">
                        <Label for="smtp_from_name">Anzeigename</Label>
                        <Input id="smtp_from_name" v-model="form.smtp_from_name" placeholder="Max Mustermann" />
                        <InputError :message="form.errors.smtp_from_name" />
                    </div>
                    <div class="grid gap-2 sm:col-span-2">
                        <Label for="smtp_from_address">Absender-E-Mail</Label>
                        <Input id="smtp_from_address" type="email" v-model="form.smtp_from_address" placeholder="Deine IServ E-Mail-Addresse" />
                        <InputError :message="form.errors.smtp_from_address" />
                    </div>
                </CardContent>
            </Card>

            <!-- AI -->
            <Card>
                <CardHeader>
                    <CardTitle>KI-Einstellungen</CardTitle>
                </CardHeader>
                <CardContent class="grid gap-4 sm:grid-cols-2">
                    <div class="grid gap-2">
                        <Label for="ai_provider">Anbieter</Label>
                        <Select v-model="form.ai_provider">
                            <SelectTrigger id="ai_provider"><SelectValue /></SelectTrigger>
                            <SelectContent>
                                <SelectItem value="gemini">Google Gemini</SelectItem>
                                <SelectItem value="cloudflare">Cloudflare Workers AI</SelectItem>
                            </SelectContent>
                        </Select>
                        <InputError :message="form.errors.ai_provider" />
                    </div>
                    <div class="grid gap-2">
                        <Label for="ai_model">Modell</Label>
                        <Input id="ai_model" v-model="form.ai_model" placeholder="gemini-1.5-flash" />
                        <InputError :message="form.errors.ai_model" />
                    </div>
                    <div class="grid gap-2 sm:col-span-2">
                        <Label for="ai_api_url">API URL (optional)</Label>
                        <Input id="ai_api_url" v-model="form.ai_api_url" placeholder="https://..." />
                        <InputError :message="form.errors.ai_api_url" />
                    </div>
                    <div class="grid gap-2 sm:col-span-2">
                        <Label for="ai_api_key">API Key{{ isEditing ? ' (leer = unverändert)' : '' }}</Label>
                        <Input id="ai_api_key" type="password" v-model="form.ai_api_key" autocomplete="new-password" />
                        <InputError :message="form.errors.ai_api_key" />
                    </div>
                </CardContent>
            </Card>

            <div class="flex items-center gap-4">
                <Button type="submit" :disabled="form.processing">
                    {{ form.processing ? 'Speichern…' : 'Konfiguration speichern' }}
                </Button>
                <span v-if="(page.props as any).flash?.success" class="text-sm text-green-600 dark:text-green-400">
                    {{ (page.props as any).flash.success }}
                </span>
            </div>
        </form>
    </div>
</template>