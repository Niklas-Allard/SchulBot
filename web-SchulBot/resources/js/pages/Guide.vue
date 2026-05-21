<script setup lang="ts">
import { Head, Link } from '@inertiajs/vue3';
import { ref, computed } from 'vue';
import { CheckCircle2, ChevronLeft, ChevronRight, BookOpen } from 'lucide-vue-next';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import StepHeader from '@/components/guide/StepHeader.vue';
import StepSection from '@/components/guide/StepSection.vue';
import ScreenshotPlaceholder from '@/components/guide/ScreenshotPlaceholder.vue';

const currentStep = ref(0);

const steps = [
    { id: 0, icon: '1', short: 'Registrierung' },
    { id: 1, icon: '2', short: 'Bot einrichten' },
    { id: 2, icon: '3', short: 'E-Mail senden' },
    { id: 3, icon: '4', short: 'Sieve-Skript' },
    { id: 4, icon: '5', short: 'Fertig!' },
];

const isFirst = computed(() => currentStep.value === 0);
const isLast  = computed(() => currentStep.value === steps.length - 1);

function next() { if (!isLast.value)  currentStep.value++; }
function prev() { if (!isFirst.value) currentStep.value--; }

const commands = [
    { tag: '#ki',        desc: 'KI-Frage stellen (ChatGPT-ähnlich)' },
    { tag: '#news',      desc: 'Aktuelle Nachrichten (Tagesschau)' },
    { tag: '#sudoku',    desc: 'Tägliches Sudoku-Rätsel' },
    { tag: '#translate', desc: 'Text übersetzen' },
    { tag: '#tasks',     desc: 'Google Tasks verwalten' },
    { tag: '#calendar',  desc: 'Google Kalender abfragen' },
    { tag: '#model',     desc: 'KI-Modell wechseln' },
    { tag: '#hilfe',     desc: 'Alle Befehle anzeigen' },
];
</script>

<template>
    <Head title="Einrichtungs-Guide – SchulBot" />

    <div class="min-h-screen bg-background text-foreground">
        <!-- Header -->
        <header class="border-b px-6 py-4">
            <div class="mx-auto flex max-w-4xl items-center">
                <div class="flex items-center gap-2 font-semibold">
                    <BookOpen class="h-5 w-5" />
                    SchulBot – Einrichtungs-Guide
                </div>
            </div>
        </header>

        <div class="mx-auto max-w-4xl px-6 py-10">

            <!-- Stepper indicators -->
            <nav class="mb-10">
                <ol class="flex items-center">
                    <li
                        v-for="(step, i) in steps"
                        :key="step.id"
                        class="flex flex-1 items-center"
                        :class="{ 'flex-none': i === steps.length - 1 }"
                    >
                        <button
                            class="flex flex-col items-center gap-1.5 text-center"
                            @click="currentStep = i"
                        >
                            <span
                                class="flex h-9 w-9 items-center justify-center rounded-full border-2 text-sm font-semibold transition-colors"
                                :class="
                                    i < currentStep
                                        ? 'border-primary bg-primary text-primary-foreground'
                                        : i === currentStep
                                            ? 'border-primary bg-primary text-primary-foreground ring-4 ring-primary/20'
                                            : 'border-muted-foreground/30 bg-background text-muted-foreground'
                                "
                            >
                                <CheckCircle2 v-if="i < currentStep" class="h-5 w-5" />
                                <span v-else>{{ step.icon }}</span>
                            </span>
                            <span
                                class="hidden text-xs font-medium sm:block"
                                :class="i === currentStep ? 'text-foreground' : 'text-muted-foreground'"
                            >
                                {{ step.short }}
                            </span>
                        </button>

                        <!-- Connector line -->
                        <div
                            v-if="i < steps.length - 1"
                            class="mx-2 h-0.5 flex-1"
                            :class="i < currentStep ? 'bg-primary' : 'bg-border'"
                        />
                    </li>
                </ol>
            </nav>

            <!-- Step content -->
            <div class="rounded-2xl border bg-card p-6 shadow-sm sm:p-8">

                <!-- ─── Step 1: Registrierung ─────────────────────────────── -->
                <div v-if="currentStep === 0">
                    <StepHeader step="1" title="Registrierung" badge="5 Min." />

                    <p class="mb-6 text-muted-foreground">
                        Erstelle einen Account auf dem SchulBot-Dashboard. Du brauchst dazu
                        deine IServ-E-Mail-Adresse (<code class="rounded bg-muted px-1 font-mono text-sm">vorname.nachname@pggv.de</code>).
                    </p>

                    <div class="space-y-8">
                        <StepSection number="a" title="Dashboard aufrufen">
                            <p>Öffne das SchulBot-Dashboard in deinem Browser und klicke auf <strong>„Registrieren"</strong>.</p>
                            <ScreenshotPlaceholder label="Screenshot: Startseite mit dem 'Registrieren'-Button (oben rechts)" />
                        </StepSection>

                        <StepSection number="b" title="Formular ausfüllen">
                            <p>
                                Gib deinen <strong>vollständigen Namen</strong>, deine
                                <strong>IServ-E-Mail</strong> (<code class="rounded bg-muted px-1 font-mono text-sm">vorname.nachname@pggv.de</code>)
                                und ein sicheres <strong>Passwort</strong> ein.
                            </p>
                            <ScreenshotPlaceholder label="Screenshot: Ausgefülltes Registrierungsformular (Passwort geschwärzt)" />
                        </StepSection>

                        <StepSection number="c" title="E-Mail-Adresse bestätigen">
                            <p>
                                Du erhältst eine Bestätigungs-E-Mail an deine IServ-Adresse.
                                Öffne sie und klicke auf den <strong>Bestätigungs-Link</strong>.
                            </p>
                            <ScreenshotPlaceholder label="Screenshot: Bestätigungs-E-Mail in IServ mit hervorgehobenem Link" />
                        </StepSection>
                    </div>
                </div>

                <!-- ─── Step 2: Bot einrichten ────────────────────────────── -->
                <div v-if="currentStep === 1">
                    <StepHeader step="2" title="Bot einrichten" badge="10 Min." />

                    <p class="mb-6 text-muted-foreground">
                        Damit der Bot <em>in deinem Namen</em> antworten kann, braucht er deine
                        IServ-SMTP-Zugangsdaten. E-Mails werden über den gemeinsamen Bot-Account
                        empfangen, aber jede Antwort kommt von <strong>deiner</strong> IServ-Adresse.
                    </p>

                    <div class="space-y-8">
                        <StepSection number="a" title="Bot-Einrichtung öffnen">
                            <p>Klicke nach dem Login in der linken Seitenleiste auf <strong>„Bot einrichten"</strong>.</p>
                            <ScreenshotPlaceholder label="Screenshot: Dashboard-Seitenleiste mit markiertem 'Bot einrichten'-Menüpunkt" />
                        </StepSection>

                        <StepSection number="b" title="IServ-SMTP-Daten herausfinden">
                            <p>Öffne in einem neuen Tab IServ und gehe zu <strong>E-Mail → Einstellungen → E-Mail-Client</strong>. Dort findest du die SMTP-Einstellungen.</p>
                            <div class="my-3 overflow-hidden rounded-lg border">
                                <table class="w-full text-sm">
                                    <thead class="bg-muted/50">
                                        <tr>
                                            <th class="px-4 py-2 text-left font-medium">Feld</th>
                                            <th class="px-4 py-2 text-left font-medium">Wert</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        <tr class="border-t"><td class="px-4 py-2 font-medium">SMTP-Server</td><td class="px-4 py-2 font-mono text-muted-foreground">pggv.de</td></tr>
                                        <tr class="border-t"><td class="px-4 py-2 font-medium">Port</td><td class="px-4 py-2 font-mono text-muted-foreground">465 (SSL)</td></tr>
                                        <tr class="border-t"><td class="px-4 py-2 font-medium">Benutzername</td><td class="px-4 py-2 font-mono text-muted-foreground">vorname.nachname <em>(ohne @pggv.de)</em></td></tr>
                                        <tr class="border-t"><td class="px-4 py-2 font-medium">Passwort</td><td class="px-4 py-2 text-muted-foreground">Dein IServ-Passwort</td></tr>
                                        <tr class="border-t"><td class="px-4 py-2 font-medium">Verschlüsselung</td><td class="px-4 py-2 font-mono text-muted-foreground">SSL/TLS</td></tr>
                                        <tr class="border-t"><td class="px-4 py-2 font-medium">Absender-E-Mail</td><td class="px-4 py-2 font-mono text-muted-foreground">vorname.nachname@pggv.de</td></tr>
                                    </tbody>
                                </table>
                            </div>
                            <ScreenshotPlaceholder label="Screenshot: IServ-Einstellungen unter E-Mail → E-Mail-Client mit den SMTP-Werten" />
                        </StepSection>

                        <StepSection number="c" title="Formular ausfüllen und speichern">
                            <p>Trage die Werte ins SchulBot-Dashboard ein und klicke auf <strong>„Konfiguration speichern"</strong>.</p>
                            <ScreenshotPlaceholder label="Screenshot: Ausgefülltes Bot-Setup-Formular im Dashboard (Passwort geschwärzt)" />
                            <ScreenshotPlaceholder label="Screenshot: Erfolgs-Meldung 'Konfiguration gespeichert' nach dem Speichern" />
                        </StepSection>

                        <StepSection number="d" title="KI-Einstellungen (optional)">
                            <p>
                                Im Abschnitt <strong>„KI-Einstellungen"</strong> kannst du wählen, welche KI du nutzen möchtest.
                                Wenn du keinen eigenen API-Key hast, lasse die Felder leer — der Bot nutzt dann die Standardkonfiguration.
                            </p>
                            <ScreenshotPlaceholder label="Screenshot: KI-Einstellungen-Bereich mit Anbieter-Auswahl (Gemini / Cloudflare)" />
                        </StepSection>
                    </div>
                </div>

                <!-- ─── Step 3: E-Mail senden ─────────────────────────────── -->
                <div v-if="currentStep === 2">
                    <StepHeader step="3" title="E-Mail an den Bot senden" badge="2 Min." />

                    <p class="mb-6 text-muted-foreground">
                        Schreibe eine E-Mail an den Bot. Du kannst jeden E-Mail-Client dafür
                        verwenden — IServ, Gmail, Outlook, etc.
                    </p>

                    <div class="space-y-8">
                        <StepSection number="a" title="Bot-Adresse">
                            <p>Sende die E-Mail an:</p>
                            <div class="my-3 flex items-center gap-3 rounded-lg border bg-muted/30 px-4 py-3">
                                <span class="font-mono text-sm font-semibold">niklas.allard.de@gmail.com</span>
                                <Badge variant="secondary">Bot-Postfach</Badge>
                            </div>
                        </StepSection>

                        <StepSection number="b" title="Betreff = Befehl">
                            <p>
                                Schreibe den gewünschten Befehl als <strong>Betreff</strong>. Das Wort nach dem
                                Tag ist deine Anfrage (Payload).
                            </p>
                            <div class="my-3 space-y-2 rounded-lg border bg-muted/30 p-4 font-mono text-sm">
                                <p><span class="text-primary">#ki</span> Was ist Photosynthese?</p>
                                <p><span class="text-primary">#news</span></p>
                                <p><span class="text-primary">#sudoku</span></p>
                                <p><span class="text-primary">#translate</span> Hallo Welt → Englisch</p>
                            </div>
                            <ScreenshotPlaceholder label="Screenshot: Neues E-Mail-Fenster in IServ mit 'niklas.allard.de@gmail.com' als Empfänger und '#ki Was ist Photosynthese?' als Betreff" />
                        </StepSection>

                        <StepSection number="c" title="Antwort erhalten">
                            <p>
                                Innerhalb von ca. <strong>30–60 Sekunden</strong> bekommst du eine
                                Antwort-E-Mail direkt in dein IServ-Postfach.
                            </p>
                            <ScreenshotPlaceholder label="Screenshot: Antwort-E-Mail vom Bot in IServ mit der KI-Antwort auf die Frage" />
                        </StepSection>

                        <StepSection number="d" title="Verfügbare Befehle">
                            <div class="grid gap-2 sm:grid-cols-2">
                                <div v-for="cmd in [
                                    { tag: '#ki',        desc: 'KI-Frage stellen (ChatGPT-ähnlich)' },
                                    { tag: '#news',      desc: 'Aktuelle Nachrichten (Tagesschau)' },
                                    { tag: '#sudoku',    desc: 'Tägliches Sudoku-Rätsel' },
                                    { tag: '#translate', desc: 'Text übersetzen' },
                                    { tag: '#tasks',     desc: 'Google Tasks verwalten' },
                                    { tag: '#calendar',  desc: 'Google Kalender abfragen' },
                                    { tag: '#model',     desc: 'KI-Modell wechseln' },
                                    { tag: '#hilfe',     desc: 'Alle Befehle anzeigen' },
                                ]" :key="cmd.tag" class="flex items-start gap-2 rounded-lg border p-3">
                                    <code class="mt-0.5 shrink-0 rounded bg-muted px-1.5 py-0.5 font-mono text-xs">{{ cmd.tag }}</code>
                                    <span class="text-sm text-muted-foreground">{{ cmd.desc }}</span>
                                </div>
                            </div>
                        </StepSection>
                    </div>
                </div>

                <!-- ─── Step 4: Sieve-Skript ──────────────────────────────── -->
                <div v-if="currentStep === 3">
                    <StepHeader step="4" title="Sieve-Skript (optional)" badge="Fortgeschritten" badgeVariant="secondary" />

                    <p class="mb-6 text-muted-foreground">
                        Mit einem Sieve-Skript in IServ kannst du dich selbst anschreiben —
                        IServ leitet E-Mails mit Bot-Tags automatisch an den Bot weiter.
                        So musst du dir die Bot-E-Mail-Adresse nicht merken.
                    </p>

                    <div class="space-y-8">
                        <StepSection number="a" title="Warum ein Sieve-Skript?">
                            <p>
                                Statt immer an <code class="rounded bg-muted px-1 font-mono text-sm">niklas.allard.de@gmail.com</code>
                                zu schreiben, kannst du einfach eine E-Mail <strong>an dich selbst</strong> schicken.
                                Das Sieve-Skript erkennt den Bot-Befehl im Betreff und leitet die Mail automatisch weiter.
                            </p>
                            <div class="my-3 rounded-lg border bg-muted/30 p-4 text-sm">
                                <p class="font-medium">Ablauf mit Sieve-Skript:</p>
                                <ol class="mt-2 list-inside list-decimal space-y-1 text-muted-foreground">
                                    <li>Du schreibst an <code class="font-mono">dich@pggv.de</code> mit Betreff <code class="font-mono">#ki Meine Frage</code></li>
                                    <li>IServ erkennt den <code class="font-mono">#ki</code>-Tag im Sieve-Filter</li>
                                    <li>IServ leitet die Mail an <code class="font-mono">niklas.allard.de@gmail.com</code> weiter</li>
                                    <li>Der Bot antwortet an deine IServ-Adresse</li>
                                </ol>
                            </div>
                        </StepSection>

                        <StepSection number="b" title="Sieve-Editor in IServ öffnen">
                            <p>
                                Gehe in IServ zu <strong>E-Mail → Einstellungen → E-Mail-Filter</strong>
                                und klicke auf <strong>„Sieve-Skript bearbeiten"</strong>.
                            </p>
                            <ScreenshotPlaceholder label="Screenshot: IServ-Navigationsmenü mit E-Mail → Einstellungen → E-Mail-Filter" />
                            <ScreenshotPlaceholder label="Screenshot: Sieve-Editor in IServ (leeres Textfeld oder mit vorhandenem Skript)" />
                        </StepSection>

                        <StepSection number="c" title="Skript einfügen">
                            <p>Kopiere das folgende Skript und füge es im Editor ein. Falls du schon ein Skript hast, füge den Teil nach dem <code class="rounded bg-muted px-1 font-mono text-sm">require</code>-Block hinzu.</p>

                            <div class="my-3 overflow-hidden rounded-lg border">
                                <div class="flex items-center justify-between border-b bg-muted/50 px-4 py-2">
                                    <span class="text-xs font-medium text-muted-foreground">sieve</span>
                                    <Badge variant="outline" class="text-xs">In Zwischenablage kopieren</Badge>
                                </div>
                                <pre class="overflow-x-auto p-4 text-sm leading-relaxed"><code>require ["redirect"];

# SchulBot – automatische Weiterleitung
# Alle E-Mails mit Bot-Befehlen im Betreff weiterleiten
if anyof (
    header :contains "subject" "#ki",
    header :contains "subject" "/ki",
    header :contains "subject" "#news",
    header :contains "subject" "#sudoku",
    header :contains "subject" "#translate",
    header :contains "subject" "#tasks",
    header :contains "subject" "#calendar",
    header :contains "subject" "#model",
    header :contains "subject" "#hilfe"
) {
    redirect "niklas.allard.de@gmail.com";
}</code></pre>
                            </div>

                            <div class="rounded-lg border border-yellow-200 bg-yellow-50 p-4 text-sm dark:border-yellow-800 dark:bg-yellow-950">
                                <p class="font-medium text-yellow-800 dark:text-yellow-200">⚠️ Hinweis</p>
                                <p class="mt-1 text-yellow-700 dark:text-yellow-300">
                                    Das Skript leitet die E-Mail weiter, <strong>löscht sie aber nicht</strong> aus deinem Postfach.
                                    Du siehst den Befehl also auch noch in deinem Posteingang.
                                </p>
                            </div>
                        </StepSection>

                        <StepSection number="d" title="Skript speichern & testen">
                            <p>Klicke auf <strong>„Speichern"</strong> im Sieve-Editor. Schreibe dann eine Test-Mail an dich selbst mit dem Betreff <code class="rounded bg-muted px-1 font-mono text-sm">#hilfe</code>.</p>
                            <ScreenshotPlaceholder label="Screenshot: Ausgefüllter Sieve-Editor mit dem Skript und markiertem 'Speichern'-Button" />
                        </StepSection>
                    </div>
                </div>

                <!-- ─── Step 5: Fertig & Testen ───────────────────────────── -->
                <div v-if="currentStep === 4">
                    <StepHeader step="5" title="Fertig! 🎉" badge="Abgeschlossen" badgeVariant="default" />

                    <p class="mb-6 text-muted-foreground">
                        Dein SchulBot ist einsatzbereit. Hier sind ein paar Dinge, die du als nächstes tun kannst.
                    </p>

                    <div class="space-y-8">
                        <StepSection number="a" title="Test-Mail senden">
                            <p>Sende eine erste Test-Mail mit dem Befehl <code class="rounded bg-muted px-1 font-mono text-sm">#hilfe</code> — du bekommst eine Liste aller verfügbaren Befehle zurück.</p>
                            <ScreenshotPlaceholder label="Screenshot: Antwort-E-Mail vom Bot mit der Hilfe-Übersicht aller Befehle" />
                        </StepSection>

                        <StepSection number="b" title="Verlauf im Dashboard prüfen">
                            <p>Öffne das Dashboard und klicke auf <strong>„Verlauf"</strong> in der Seitenleiste. Dort siehst du alle verarbeiteten Befehle mit Antwort.</p>
                            <ScreenshotPlaceholder label="Screenshot: Verlaufs-Seite im Dashboard mit dem ersten Befehl und der Bot-Antwort" />
                        </StepSection>

                        <StepSection number="c" title="KI-Frage ausprobieren">
                            <p>Probiere eine echte KI-Frage aus:</p>
                            <div class="my-3 rounded-lg border bg-muted/30 p-4 font-mono text-sm">
                                <p><span class="text-muted-foreground">An:</span> niklas.allard.de@gmail.com</p>
                                <p><span class="text-muted-foreground">Betreff:</span> <span class="text-primary">#ki</span> Erkläre mir die Relativitätstheorie in 3 Sätzen</p>
                            </div>
                            <ScreenshotPlaceholder label="Screenshot: KI-Antwort vom Bot in IServ mit der Erklärung der Relativitätstheorie" />
                        </StepSection>

                        <div class="mt-6 flex flex-col gap-3 rounded-xl border bg-muted/30 p-6 sm:flex-row sm:items-center sm:justify-between">
                            <div>
                                <p class="font-semibold">Bereit loszulegen?</p>
                                <p class="text-sm text-muted-foreground">Erstelle deinen Account oder melde dich an.</p>
                            </div>
                            <div class="flex gap-3">
                                <Button as-child variant="outline">
                                    <Link href="/login">Anmelden</Link>
                                </Button>
                                <Button as-child>
                                    <Link href="/register">Jetzt registrieren</Link>
                                </Button>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Navigation buttons -->
                <div class="mt-10 flex items-center justify-between border-t pt-6">
                    <Button variant="outline" :disabled="isFirst" @click="prev">
                        <ChevronLeft class="mr-1 h-4 w-4" />
                        Zurück
                    </Button>
                    <span class="text-sm text-muted-foreground">
                        Schritt {{ currentStep + 1 }} von {{ steps.length }}
                    </span>
                    <Button v-if="!isLast" @click="next">
                        Weiter
                        <ChevronRight class="ml-1 h-4 w-4" />
                    </Button>
                    <Button v-else as-child>
                        <Link href="/register">Jetzt registrieren</Link>
                    </Button>
                </div>
            </div>

        </div>
    </div>
</template>

<!-- ─── Sub-components as inline templates ───────────────────────────────── -->