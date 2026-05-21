<?php

namespace App\Http\Controllers;

use App\Models\BotConfig;
use Illuminate\Http\RedirectResponse;
use Illuminate\Http\Request;
use Inertia\Inertia;
use Inertia\Response;

class BotConfigController extends Controller
{
    public function edit(Request $request): Response
    {
        $config = $request->user()->botConfig;

        return Inertia::render('bot/Setup', [
            'config' => $config ? [
                'imap_host'         => $config->imap_host,
                'imap_port'         => $config->imap_port,
                'imap_username'     => $config->imap_username,
                'imap_security'     => $config->imap_security,
                'imap_mailbox'      => $config->imap_mailbox,
                'smtp_host'         => $config->smtp_host,
                'smtp_port'         => $config->smtp_port,
                'smtp_username'     => $config->smtp_username,
                'smtp_security'     => $config->smtp_security,
                'smtp_from_name'    => $config->smtp_from_name,
                'smtp_from_address' => $config->smtp_from_address,
                'ai_provider'       => $config->ai_provider,
                'ai_api_url'        => $config->ai_api_url,
                'ai_model'          => $config->ai_model,
                'is_active'         => $config->is_active,
                // passwords are never sent to frontend
            ] : null,
        ]);
    }

    public function update(Request $request): RedirectResponse
    {
        $data = $request->validate([
            'imap_host'         => ['required', 'string', 'max:255'],
            'imap_port'         => ['required', 'integer', 'min:1', 'max:65535'],
            'imap_username'     => ['required', 'string', 'max:255'],
            'imap_password'     => ['nullable', 'string', 'max:255'],
            'imap_security'     => ['required', 'in:SSL,STARTTLS'],
            'imap_mailbox'      => ['required', 'string', 'max:64'],
            'smtp_host'         => ['required', 'string', 'max:255'],
            'smtp_port'         => ['required', 'integer', 'min:1', 'max:65535'],
            'smtp_username'     => ['required', 'string', 'max:255'],
            'smtp_password'     => ['nullable', 'string', 'max:255'],
            'smtp_security'     => ['required', 'in:SSL,STARTTLS'],
            'smtp_from_name'    => ['required', 'string', 'max:255'],
            'smtp_from_address' => ['required', 'email', 'max:255'],
            'ai_provider'       => ['required', 'in:gemini,cloudflare'],
            'ai_api_url'        => ['nullable', 'string', 'max:512'],
            'ai_api_key'        => ['nullable', 'string', 'max:255'],
            'ai_model'          => ['nullable', 'string', 'max:128'],
            'is_active'         => ['boolean'],
        ]);

        $config = $request->user()->botConfig;

        if ($config) {
            $fillable = $data;
            // Only update passwords if a new value was provided
            if (empty($fillable['imap_password'])) {
                unset($fillable['imap_password']);
            }
            if (empty($fillable['smtp_password'])) {
                unset($fillable['smtp_password']);
            }
            if (empty($fillable['ai_api_key'])) {
                unset($fillable['ai_api_key']);
            }
            $config->update($fillable);
        } else {
            $request->user()->botConfig()->create($data);
        }

        return back()->with('success', 'Bot-Konfiguration gespeichert.');
    }
}